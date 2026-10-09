package io.github.maas3n.mattmux

import android.content.ContentResolver
import android.net.Uri
import java.io.IOException

/**
 * Native Blu-ray ISO inspection via a seekable SAF ParcelFileDescriptor.
 * Reads at most 6144 bytes through libbluray; this does not yet create MKVs.
 */
internal class BlurayNativeIsoBridge(private val resolver: ContentResolver) {
    internal data class Probe(
        val playlist: Int,
        val duration90kHz: Long,
        val bytesRead: Int,
    )

    companion object {
        private val loadError = runCatching {
            System.loadLibrary("muksmatt_bluray_udfread")
            System.loadLibrary("avutil")
            System.loadLibrary("swresample")
            System.loadLibrary("avcodec")
            System.loadLibrary("avformat")
            System.loadLibrary("bluray")
            System.loadLibrary("muksmatt_bluray")
        }.exceptionOrNull()

        internal fun parseNavigation(payload: String): Probe {
            val lines = payload.trimEnd('\n').split('\n')
            require(lines.size == 3 && lines[0] == "MUKSMATT_ANDROID_BD_1") {
                "Unexpected native Blu-ray probe result"
            }
            val fields = lines[1].split('\t')
            val readFields = lines[2].split('\t')
            require(fields.size == 3 && fields[0] == "P" &&
                readFields.size == 2 && readFields[0] == "R") {
                "Invalid native Blu-ray probe fields"
            }
            fun number(raw: String): Long {
                require(raw.isNotEmpty() && raw.length <= 16 && raw.all { it in '0'..'9' }) {
                    "Invalid native Blu-ray integer"
                }
                return raw.toLongOrNull() ?: throw IllegalArgumentException("Blu-ray integer out of range")
            }
            val id = number(fields[1])
            val duration = number(fields[2])
            val read = number(readFields[1])
            require(id in 0..99999 && duration in 1..(90_000L * 3600 * 48) &&
                read in 0..6144) {
                "Invalid native Blu-ray playlist, duration or read count"
            }
            return Probe(id.toInt(), duration, read.toInt())
        }
    }

    private external fun nativeInspectIso(
        fd: Int,
        requestedPlaylist: Int,
        sampleBytes: Int,
    ): String

    private external fun nativeRemuxIso(
        sourceFd: Int, outputFd: Int, playlist: Int,
        selectedStreamIndexes: IntArray?, includeChapters: Boolean,
    ): String?

    /**
     * Direct Blu-ray ISO-to-MKV remux. All eligible streams are copied except
     * Blu-ray LPCM, which must always be encoded losslessly to FLAC.
     * Native playlist selection is independent of the existing DVD engine.
     */
    fun remux(
        isoUri: Uri,
        outputTreeUri: Uri,
        outputName: String,
        playlist: Int? = null,
        selectedStreamIndexes: IntArray? = null,
        includeChapters: Boolean = true,
    ): Uri {
        require(playlist == null || playlist in 0..99999) { "Invalid Blu-ray playlist" }
        require(selectedStreamIndexes == null ||
            (selectedStreamIndexes.isNotEmpty() &&
             selectedStreamIndexes.size <= 256 &&
             selectedStreamIndexes.all { it >= 0 } &&
             selectedStreamIndexes.toSet().size == selectedStreamIndexes.size)) {
            "Invalid Blu-ray stream selection"
        }
        if (!isAvailable) throw IOException(unavailableReason ?: "Blu-ray runtime unavailable")
        val source = resolver.openFileDescriptor(isoUri, "r")
            ?: throw IOException("SAF provider cannot open Blu-ray ISO")
        return source.use {
            BlurayMkvSafOutput.create(resolver, outputTreeUri, outputName) { outputFd ->
                val error = nativeRemuxIso(
                    it.fd, outputFd, playlist ?: -1,
                    selectedStreamIndexes, includeChapters
                )
                if (error != null) throw IOException(error)
            }
        }
    }

    val isAvailable: Boolean get() = loadError == null
    val unavailableReason: String?
        get() = loadError?.let { "Native Blu-ray reader unavailable: " + it.javaClass.simpleName }

    fun inspect(
        isoUri: Uri,
        playlist: Int? = null,
        sampleBytes: Int = 6144,
    ): Probe {
        require(playlist == null || playlist in 0..99999) { "Invalid Blu-ray playlist" }
        require(sampleBytes in 0..6144) { "Blu-ray probe must read at most 6144 bytes" }
        if (!isAvailable) throw IOException(unavailableReason ?: "Native libbluray is unavailable")
        val descriptor = resolver.openFileDescriptor(isoUri, "r")
            ?: throw IOException("SAF provider could not open the Blu-ray ISO")
        return descriptor.use {
            parseNavigation(nativeInspectIso(it.fd, playlist ?: -1, sampleBytes))
        }
    }
}
