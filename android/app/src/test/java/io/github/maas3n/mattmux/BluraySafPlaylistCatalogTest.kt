package io.github.maas3n.mattmux

import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Test

class BluraySafPlaylistCatalogTest {
    private fun mpls(seconds: Long): ByteArray {
        val data = ByteArray(256)
        "MPLS0300".toByteArray(Charsets.US_ASCII).copyInto(data)
        fun put16(offset: Int, value: Int) {
            data[offset] = (value shr 8).toByte()
            data[offset + 1] = value.toByte()
        }
        fun put32(offset: Int, value: Long) {
            for (i in 0..3) data[offset + i] = (value shr (8 * (3 - i))).toByte()
        }
        put32(8, 32)
        put32(12, 128)
        put32(32, 32)
        put16(38, 1)
        put16(42, 20)
        put32(56, 0)
        put32(60, seconds * 45_000)
        put32(128, 16)
        put16(132, 1)
        data[135] = 1
        put16(136, 0)
        put32(138, 0)
        return data
    }

    private class FakeProvider(
        private val files: Map<String, ByteArray> = mapOf(
            "short" to ByteArray(0),
        ),
    ) : BlurayDocumentProvider {
        private val root = listOf(BlurayDocumentEntry("bdmv", "bdmv", true))
        private val bdmv = listOf(BlurayDocumentEntry("playlist", "PLAYLIST", true))
        private val playlists = listOf(
            BlurayDocumentEntry("short", "00002.mpls", false),
            BlurayDocumentEntry("long", "00800.MPLS", false),
            BlurayDocumentEntry("ignored", "index.bdmv", false),
        )
        override fun listChildren(parentDocumentId: String) = when (parentDocumentId) {
            "root" -> root
            "bdmv" -> bdmv
            "playlist" -> playlists
            else -> emptyList()
        }
        override fun readPlaylist(documentId: String) =
            files[documentId] ?: throw IllegalArgumentException("Unrecognized document ID")
    }

    @Test fun scansSelectedTreeAndChoosesLongestPlaylist() {
        val provider = FakeProvider(mapOf("short" to mpls(10), "long" to mpls(90)))
        val playlists = BluraySafPlaylistCatalog.scan(
            BlurayDocumentEntry("root", "Feature disc", true),
            provider
        )
        assertEquals(listOf(2, 800), playlists.map { it.number })
        assertEquals(90L * 45000, BluraySafPlaylistCatalog.select(playlists).navigation.durationTicks)
        assertEquals(2, BluraySafPlaylistCatalog.select(playlists, 2).number)
    }

    @Test fun invalidNavigationAndPlaylistAreRejected() {
        val provider = FakeProvider(mapOf("short" to mpls(10), "long" to byteArrayOf(1, 2, 3)))
        assertThrows(IllegalArgumentException::class.java) {
            BluraySafPlaylistCatalog.scan(BlurayDocumentEntry("root", "Disc", true), provider)
        }
        val good = FakeProvider(mapOf("short" to mpls(10), "long" to mpls(20)))
        val list = BluraySafPlaylistCatalog.scan(BlurayDocumentEntry("root", "Disc", true), good)
        assertThrows(IllegalArgumentException::class.java) {
            BluraySafPlaylistCatalog.select(list, 99999)
        }
        assertThrows(IllegalArgumentException::class.java) {
            BluraySafPlaylistCatalog.scan(BlurayDocumentEntry("root", "Disc.iso", false), good)
        }
    }
}
