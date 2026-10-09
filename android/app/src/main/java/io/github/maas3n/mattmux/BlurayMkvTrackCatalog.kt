package io.github.maas3n.mattmux

/**
 * Native Blu-ray stream selectors use the FFmpeg MPEG-TS stream index.
 * Discover tracks immediately before choosing indices, because MPLS track
 * metadata and MPEG-TS stream ordering are not interchangeable.
 */
internal data class BlurayMkvTrack(val index: Int, val type: String, val codec: String) {
    val requiresFlac: Boolean get() = codec == "pcm_bluray"
}

internal object BlurayMkvTrackCatalog {
    private val codecName = Regex("[a-zA-Z0-9_]{1,64}")

    fun parse(value: String): List<BlurayMkvTrack> {
        val lines = value.trimEnd('\n').split('\n')
        require(lines.isNotEmpty() && lines[0] == "MUKSMATT_BD_TRACKS_1" &&
            lines.size in 2..257) { "Invalid native Blu-ray track catalog" }
        val seen = HashSet<Int>()
        return lines.drop(1).map { line ->
            val cols = line.split('\t')
            require(cols.size == 4 && cols[0] == "S" &&
                cols[1].isNotBlank() && cols[1].all { it in '0'..'9' }) {
                "Invalid Blu-ray stream record"
            }
            val index = cols[1].toIntOrNull()
                ?: throw IllegalArgumentException("Invalid Blu-ray stream index")
            require(index in 0..255 && seen.add(index) &&
                cols[2] in setOf("video", "audio", "subtitle") &&
                codecName.matches(cols[3])) {
                "Invalid or duplicate Blu-ray track"
            }
            BlurayMkvTrack(index, cols[2], cols[3])
        }
    }

    fun validateSelection(
        tracks: List<BlurayMkvTrack>,
        selected: IntArray?,
    ): IntArray? {
        if (selected == null) return null
        require(selected.isNotEmpty() && selected.size <= 256 &&
            selected.toSet().size == selected.size) { "Select at least one unique track" }
        val valid = tracks.map { it.index }.toSet()
        require(selected.all { it in valid }) {
            "Blu-ray stream selection no longer matches the current playlist"
        }
        return selected.copyOf()
    }
}
