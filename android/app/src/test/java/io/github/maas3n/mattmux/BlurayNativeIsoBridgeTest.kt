package io.github.maas3n.mattmux

import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Test

class BlurayNativeIsoBridgeTest {
    @Test fun acceptsVersionedNativeResult() {
        val result = BlurayNativeIsoBridge.parseNavigation(
            "MUKSMATT_ANDROID_BD_1\nP\t800\t5400000\nR\t6144\n"
        )
        assertEquals(800, result.playlist)
        assertEquals(5400000L, result.duration90kHz)
        assertEquals(6144, result.bytesRead)
    }

    @Test fun rejectsCorruptOrUntrustedNativeOutput() {
        val bad = listOf(
            "",
            "MUKSMATT_ANDROID_BD_0\nP\t800\t1\nR\t1\n",
            "MUKSMATT_ANDROID_BD_1\nP\t800\t1\n",
            "MUKSMATT_ANDROID_BD_1\nP\t-1\t1\nR\t1\n",
            "MUKSMATT_ANDROID_BD_1\nP\t100000\t1\nR\t1\n",
            "MUKSMATT_ANDROID_BD_1\nP\t800\t0\nR\t1\n",
            "MUKSMATT_ANDROID_BD_1\nP\t800\t1\nR\t6145\n",
            "MUKSMATT_ANDROID_BD_1\nP\t800\t1\nR\t1\nEXTRA\n",
            "MUKSMATT_ANDROID_BD_1\nP\t800\t999999999999999999999\nR\t1\n",
        )
        bad.forEach { value ->
            assertThrows(IllegalArgumentException::class.java) {
                BlurayNativeIsoBridge.parseNavigation(value)
            }
        }
    }
}
