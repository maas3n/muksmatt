package io.github.maas3n.mattmux

import android.app.Activity
import android.app.Instrumentation
import android.os.Bundle
import android.provider.DocumentsContract
import android.util.Log
import java.io.File
import java.io.IOException

/**
 * Debug-only Android device integration using genuine SAF callbacks.
 * Unencrypted, self-authored Blu-ray fixtures; NOT a physical-disc test.
 */
class BluraySmokeInstrumentation : Instrumentation() {
    private val authority = "io.github.maas3n.muksmatt.demux-test"

    override fun onCreate(arguments: Bundle?) {
        super.onCreate(arguments)
        start()
    }

    private fun unpack(source: String, target: File) {
        val names = targetContext.assets.list(source).orEmpty()
        if (names.isEmpty()) {
            target.parentFile?.mkdirs()
            targetContext.assets.open(source).use { data ->
                target.outputStream().use { output -> data.copyTo(output) }
            }
        } else {
            check(target.mkdirs() || target.isDirectory)
            names.forEach { name -> unpack("$source/$name", File(target, name)) }
        }
    }

    private fun tracks(path: File): Pair<List<String>, Boolean> {
        check(path.isFile && path.length() > 0) { "Empty native MKV: $path" }
        val records = AdvancedMergerNative().probe(path.absolutePath).toList()
        val kinds = records.map { it.split('\t') }
        val codecs = kinds.filter { it.size > 2 && it[0] != "-1" }.map { it[2] }
        return codecs to kinds.any { it.size > 1 && it[1] == "chapters" }
    }

    private fun ensureMkv(
        safRoot: File, outputUri: android.net.Uri, filename: String,
        wantVideo: Boolean, wantAudio: Boolean, chapters: Boolean,
    ) {
        val id = DocumentsContract.getDocumentId(outputUri)
        val output = File(safRoot, id)
        check(output.name == filename && output.isFile && output.length() > 0) {
            "Blu-ray SAF did not publish expected MKV: $output"
        }
        val (codecs, hasChapters) = tracks(output)
        check(codecs.contains("mpeg2video") == wantVideo && codecs.contains("flac") == wantAudio) {
            "Wrong Blu-ray output streams in $filename: $codecs"
        }
        check(hasChapters == chapters) { "Blu-ray chapter selection lost for $filename" }
        check(safRoot.walkTopDown().none {
            it.isFile && (it.name.contains(".partial-") || it.name == "source.mkv")
        }) { "Blu-ray created or leaked an intermediate MKV" }
        Log.i("muKsMaTTBlurayTest",
            "Device SAF MKV $filename: bytes=${output.length()} streams=$codecs chapters=$hasChapters")
    }

    override fun onStart() {
        val result = Bundle()
        val safRoot = File(targetContext.cacheDir, "dvd-saf-test")
        try {
            check(safRoot.deleteRecursively())
            check(safRoot.mkdirs())
            unpack("bluray-smoke/BDMV", File(safRoot, "bluray-disc/BDMV"))
            if (!targetContext.assets.list("bluray-smoke/CERTIFICATE").isNullOrEmpty()) {
                unpack("bluray-smoke/CERTIFICATE", File(safRoot, "bluray-disc/CERTIFICATE"))
            }
            unpack("bluray-smoke/movie.iso", File(safRoot, "movie.iso"))
            check(File(safRoot, "bluray-disc/BDMV/PLAYLIST").isDirectory)
            check(File(safRoot, "bluray-disc/BDMV/STREAM").isDirectory)
            File(safRoot, "bluray-output").mkdirs()
            val source = DocumentsContract.buildTreeDocumentUri(authority, "bluray-disc")
            val destination = DocumentsContract.buildTreeDocumentUri(authority, "bluray-output")
            val isoUri = DocumentsContract.buildDocumentUri(authority, "movie.iso")
            val tree = BluraySafTreeBridge(targetContext.contentResolver, source)
            val iso = BlurayNativeIsoBridge(targetContext.contentResolver)
            check(iso.isAvailable) { iso.unavailableReason ?: "No ISO runtime" }
            val folderTracks = tree.probeStreams()
            val imageTracks = iso.probeStreams(isoUri)
            check(folderTracks.any { it.type == "video" && it.codec == "mpeg2video" })
            check(folderTracks.any { it.type == "audio" && it.requiresFlac })
            check(imageTracks.map { it.type to it.codec } == folderTracks.map { it.type to it.codec }) {
                "SAF folder and ISO probe found different Blu-ray tracks"
            }
            check(tree.inspect(sampleBytes = 6144).bytesRead > 0)
            check(iso.inspect(isoUri, sampleBytes = 6144).bytesRead > 0)
            ensureMkv(safRoot, tree.remux(destination, "folder-full.mkv"),
                      "folder-full.mkv", true, true, true)
            ensureMkv(safRoot, iso.remux(isoUri, destination, "iso-full.mkv"),
                      "iso-full.mkv", true, true, true)
            val audio = folderTracks.filter { it.type == "audio" }.map { it.index }.toIntArray()
            val video = folderTracks.filter { it.type == "video" }.map { it.index }.toIntArray()
            check(audio.isNotEmpty() && video.isNotEmpty())
            ensureMkv(safRoot,
                tree.remux(destination, "audio-only.mkv", selectedStreamIndexes = audio,
                           includeChapters = false),
                "audio-only.mkv", false, true, false)
            ensureMkv(safRoot,
                iso.remux(isoUri, destination, "video-only.mkv",
                          selectedStreamIndexes = video, includeChapters = false),
                "video-only.mkv", true, false, false)

            // An interrupted write must not publish an incomplete document.
            try {
                BlurayMkvSafOutput.create(targetContext.contentResolver, destination, "abort.mkv") {
                    throw IOException("Injected output failure")
                }
                error("Injected SAF output failure was silently ignored")
            } catch (expected: IOException) {
                check(expected.message == "Injected output failure") { expected.toString() }
            }
            check(!File(safRoot, "bluray-output/abort.mkv").exists())
            check(File(safRoot, "bluray-output").listFiles().orEmpty().none {
                it.name.contains(".partial-")
            }) { "Failed Blu-ray SAF remux left partial output" }
            result.putString("stream", "MUKSMATT_BLURAY_SAF_PASS")
            finish(Activity.RESULT_OK, result)
        } catch (failure: Throwable) {
            Log.e("muKsMaTTBlurayTest", "Android Blu-ray device test failed", failure)
            result.putString("stream", failure.stackTraceToString())
            finish(Activity.RESULT_CANCELED, result)
        } finally {
            safRoot.deleteRecursively()
        }
    }
}
