package io.github.maas3n.mattmux

/**
 * Bounded MPLS navigation parsing for Android/ChromeOS SAF-based Blu-ray sources.
 * This does not open encrypted media; it accepts bytes obtained by a future
 * read-only DocumentsProvider/SAF playlist reader.
 *
 * Equivalent 45 kHz PlayItem/PlayListMark semantics to the desktop Go parser.
 * Keeping the existing Kotlin/JNI namespace avoids disturbing the DVD engine.
 */
internal data class BlurayChapterMark(
    val number: Int,
    val startTicks: Long,
    val endTicks: Long,
)

internal data class BlurayPlaylistNavigation(
    val durationTicks: Long,
    val clipCount: Int,
    val chapters: List<BlurayChapterMark>,
)

internal object BlurayMplsNavigation {
    private const val MAX_BYTES = 16 * 1024 * 1024
    private const val MAX_TICKS = 45_000L * 60 * 60 * 48

    fun parse(bytes: ByteArray): BlurayPlaylistNavigation {
        require(bytes.size in 20..MAX_BYTES) { "Invalid MPLS size" }
        val signature = bytes.copyOfRange(0, 8).toString(Charsets.US_ASCII)
        require(signature in setOf("MPLS0100", "MPLS0200", "MPLS0240", "MPLS0300")) {
            "Invalid MPLS header"
        }
        fun inside(offset: Long, count: Long): Boolean =
            offset >= 0 && count >= 0 && offset <= bytes.size.toLong() &&
                count <= bytes.size.toLong() - offset
        fun u16(offset: Long): Int {
            require(inside(offset, 2)) { "Truncated MPLS field" }
            val at = offset.toInt()
            return ((bytes[at].toInt() and 255) shl 8) or (bytes[at + 1].toInt() and 255)
        }
        fun u32(offset: Long): Long {
            require(inside(offset, 4)) { "Truncated MPLS field" }
            val at = offset.toInt()
            return ((bytes[at].toLong() and 255) shl 24) or
                ((bytes[at + 1].toLong() and 255) shl 16) or
                ((bytes[at + 2].toLong() and 255) shl 8) or
                (bytes[at + 3].toLong() and 255)
        }
        fun ticks(offset: Long): Long = u32(offset) and 0x7fff_ffffL

        val playlistStart = u32(8)
        val marksStart = u32(12)
        require(inside(playlistStart, 10) && inside(marksStart, 6)) { "Invalid MPLS section offsets" }
        val playlistLength = u32(playlistStart) + 4
        val marksLength = u32(marksStart) + 4
        require(playlistLength >= 10 && marksLength >= 6 &&
            inside(playlistStart, playlistLength) && inside(marksStart, marksLength)) {
            "Truncated MPLS sections"
        }
        val playlistEnd = playlistStart + playlistLength
        val marksEnd = marksStart + marksLength

        data class Clip(val start: Long, val end: Long, val timeline: Long)
        val itemCount = u16(playlistStart + 6)
        require(itemCount > 0) { "MPLS has no PlayItems" }
        val clips = ArrayList<Clip>(itemCount)
        var offset = playlistStart + 10
        var duration = 0L
        repeat(itemCount) {
            require(offset + 22 <= playlistEnd) { "Truncated MPLS PlayItem" }
            val length = u16(offset).toLong()
            require(length >= 20 && offset + 2 + length <= playlistEnd) {
                "Invalid MPLS PlayItem length"
            }
            val start = ticks(offset + 14)
            val end = ticks(offset + 18)
            require(end > start) { "Invalid MPLS PlayItem clock" }
            val span = end - start
            require(duration + span <= MAX_TICKS) { "MPLS longer than 48 hours" }
            clips += Clip(start, end, duration)
            duration += span
            offset += 2 + length
        }

        val markCount = u16(marksStart + 4)
        var markPos = marksStart + 6
        require(markCount.toLong() <= (marksEnd - markPos) / 14) {
            "Truncated MPLS chapter marks"
        }
        val starts = sortedSetOf<Long>()
        repeat(markCount) {
            val type = bytes[(markPos + 1).toInt()].toInt() and 255
            val index = u16(markPos + 2)
            val timestamp = ticks(markPos + 4)
            if (type == 1 && index < clips.size) {
                val clip = clips[index]
                if (timestamp >= clip.start && timestamp < clip.end) {
                    val point = clip.timeline + timestamp - clip.start
                    if (point >= 0 && duration - point > 45_000) starts += point
                }
            }
            markPos += 14
        }
        val marks = starts.toList()
        val chapters = marks.mapIndexedNotNull { index, start ->
            val end = marks.getOrNull(index + 1) ?: duration
            if (end <= start) null else BlurayChapterMark(index + 1, start, end)
        }
        return BlurayPlaylistNavigation(duration, clips.size, chapters)
    }
}
