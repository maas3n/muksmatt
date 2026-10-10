package io.github.maas3n.mattmux

import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Assert.assertTrue
import org.junit.Test

class BlurayMkvTrackCatalogTest {
    private val catalog = "MUKSMATT_BD_TRACKS_1\n" +
        "S\t0\tvideo\thevc\n" +
        "S\t1\taudio\tpcm_bluray\n" +
        "S\t2\taudio\ttruehd\n" +
        "S\t3\tsubtitle\thdmv_pgs_subtitle\n"

    @Test fun parsesActualMpegTsIndicesAndLpcmPolicy() {
        val tracks = BlurayMkvTrackCatalog.parse(catalog)
        assertEquals(4, tracks.size)
        assertEquals(listOf(0, 1, 2, 3), tracks.map { it.index })
        assertTrue(tracks[1].requiresFlac)
        assertEquals(false, tracks[2].requiresFlac)
        val selected = BlurayMkvTrackCatalog.validateSelection(tracks, intArrayOf(0, 1, 3))
        assertEquals(listOf(0, 1, 3), selected!!.toList())
    }

    @Test fun refusesMalformedCatalogAndInvalidSelections() {
        val bad = listOf(
            "",
            "MUKSMATT_BD_TRACKS_0\nS\t0\tvideo\thevc\n",
            "MUKSMATT_BD_TRACKS_1\nS\t-1\tvideo\thevc\n",
            "MUKSMATT_BD_TRACKS_1\nS\t256\tvideo\thevc\n",
            "MUKSMATT_BD_TRACKS_1\nS\t0\tvideo\thevc\nS\t0\taudio\taac\n",
            "MUKSMATT_BD_TRACKS_1\nS\t0\texecutable\thevc\n",
            "MUKSMATT_BD_TRACKS_1\nS\t0\tvideo\thevc\tunsafe\n",
            "MUKSMATT_BD_TRACKS_1\nS\t0\tvideo\thevc/../../x\n",
            "MUKSMATT_BD_TRACKS_1\n",
        )
        bad.forEach { value ->
            assertThrows(IllegalArgumentException::class.java) {
                BlurayMkvTrackCatalog.parse(value)
            }
        }
        val tracks = BlurayMkvTrackCatalog.parse(catalog)
        listOf(intArrayOf(), intArrayOf(0, 0), intArrayOf(-1), intArrayOf(99)).forEach { indices ->
            assertThrows(IllegalArgumentException::class.java) {
                BlurayMkvTrackCatalog.validateSelection(tracks, indices)
            }
        }
    }
}
