package io.github.maas3n.mattmux

import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Test

class BlurayMplsNavigationTest {
    private fun fixture(duplicate: Boolean = false): ByteArray {
        val b = ByteArray(512)
        "MPLS0300".toByteArray(Charsets.US_ASCII).copyInto(b)
        fun write16(at: Int, value: Int) {
            b[at] = (value shr 8).toByte()
            b[at + 1] = value.toByte()
        }
        fun write32(at: Int, value: Long) {
            for (i in 0..3) b[at + i] = (value shr (8 * (3 - i))).toByte()
        }
        write32(8, 64)
        write32(12, 256)
        write32(64, 54)
        write16(70, 2)
        write16(74, 20)
        write32(88, 0)
        write32(92, 40 * 45000L)
        write16(96, 20)
        write32(110, 100 * 45000L)
        write32(114, 120 * 45000L)
        write32(256, 2 + 14L * if (duplicate) 3 else 2)
        write16(260, if (duplicate) 3 else 2)
        b[263] = 1
        write16(264, 0)
        write32(266, 0)
        b[277] = 1
        write16(278, 1)
        write32(280, 100 * 45000L)
        if (duplicate) {
            b[291] = 1
            write16(292, 1)
            write32(294, 100 * 45000L)
        }
        return b
    }

    @Test fun parsesMultiClipMplsWithoutClockResetChapterDrift() {
        val p = BlurayMplsNavigation.parse(fixture())
        assertEquals(2, p.clipCount)
        assertEquals(60 * 45000L, p.durationTicks)
        assertEquals(2, p.chapters.size)
        assertEquals(0L, p.chapters[0].startTicks)
        assertEquals(40 * 45000L, p.chapters[0].endTicks)
        assertEquals(40 * 45000L, p.chapters[1].startTicks)
        assertEquals(60 * 45000L, p.chapters[1].endTicks)
    }

    @Test fun duplicateMarksAreNotDuplicated() {
        val parsed = BlurayMplsNavigation.parse(fixture(duplicate = true))
        assertEquals(2, parsed.chapters.size)
    }

    @Test fun rejectsMalformedNavigation() {
        val short = byteArrayOf(77, 80, 76, 83)
        val invalid = fixture().also { it[0] = 'X'.code.toByte() }
        val truncated = fixture().also { it[12] = 127 }
        for (bytes in listOf(short, invalid, truncated)) {
            assertThrows(IllegalArgumentException::class.java) {
                BlurayMplsNavigation.parse(bytes)
            }
        }
    }
}
