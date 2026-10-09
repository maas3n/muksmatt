package io.github.maas3n.mattmux

import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertThrows
import org.junit.Test

class BluraySafPathIndexTest {
    private fun sample(): BluraySafPathIndex = BluraySafPathIndex().apply {
        register("", "root-doc", "", true)
        register("BDMV", "bdmv-doc", "BDMV", true)
        register("BDMV/PLAYLIST", "playlist-doc", "PLAYLIST", true)
        register("BDMV/PLAYLIST/00800.mpls", "playlist-file", "00800.mpls", false)
        register("BDMV/STREAM", "stream-doc", "STREAM", true)
        register("BDMV/STREAM/00001.m2ts", "stream-file", "00001.m2ts", false)
    }

    @Test fun acceptsCanonicalCaseInsensitiveDiscPaths() {
        val i = sample()
        assertEquals("playlist-file", i.node("/bdmv/playlist/00800.MPLS")?.documentId)
        assertEquals("stream-file", i.node("BDMV/stream/00001.m2ts")?.documentId)
        assertArrayEquals(arrayOf("00800.mpls"), i.names("bdmv/playlist"))
        assertNull(i.names("bdmv/stream/00001.m2ts"))
        assertNull(i.names("BDMV/PLAYLIST/UNKNOWN"))
        assertArrayEquals(arrayOf("PLAYLIST", "STREAM"), i.names("/BDMV"))
    }

    @Test fun rejectsTraversalAndSpecialPaths() {
        val bad = listOf(
            "../BDMV", "BDMV/../AACS", "BDMV//STREAM", "BDMV/./STREAM",
            "BDMV\\STREAM", "BDMV/STREAM/", "BDMV/STREAM/evil\u0000.m2ts",
            "BDMV/STREAM/ümlaut.m2ts", "BDMV/?", "/" + "A".repeat(1200),
        )
        bad.forEach { path ->
            assertNull("Must reject: " + path, BluraySafPathIndex.canonical(path))
            assertNull(sample().node(path))
        }
    }

    @Test fun refusesAmbiguousCaseAndInvalidHierarchy() {
        val i = sample()
        assertThrows(IllegalArgumentException::class.java) {
            i.register("bdmv/playlist/00800.MPLS", "other-doc", "00800.MPLS", false)
        }
        assertThrows(IllegalArgumentException::class.java) {
            i.register("BDMV/NOT_THERE/file", "missing", "file", false)
        }
        assertThrows(IllegalArgumentException::class.java) {
            i.register("BDMV/STREAM/00001.m2ts/child", "wrong", "child", false)
        }
        assertThrows(IllegalArgumentException::class.java) {
            i.register("BDMV/STREAM/../../payload", "wrong", "payload", false)
        }
    }

    @Test fun supportsBdmvRootSelectedViaSaf() {
        val i = BluraySafPathIndex()
        i.register("", "bdmv-document", "", true)
        i.register("BDMV", "bdmv-document", "BDMV", true)
        i.register("BDMV/STREAM", "stream", "STREAM", true)
        assertEquals("bdmv-document", i.node("BDMV")?.documentId)
    }
}
