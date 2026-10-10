package io.github.maas3n.mattmux

import android.app.Activity
import android.app.Instrumentation
import android.os.Bundle
import android.provider.DocumentsContract
import android.util.Log
import java.io.IOException

/**
 * CI-only Android device integration using genuine SAF callbacks.
 * Unencrypted, self-authored Blu-ray fixtures; NOT a physical-disc test.
 */
class BluraySmokeInstrumentation : Instrumentation() {
    private val authority get() = context.packageName + ".demux-test"
    private val resolver get() = targetContext.contentResolver
    private val providerUri get() = android.net.Uri.parse("content://$authority")

    override fun onCreate(arguments: Bundle?) {
        super.onCreate(arguments)
        start()
    }

    private fun outputNames(): List<String> {
        val tree = DocumentsContract.buildTreeDocumentUri(authority, "bluray-output")
        val children = DocumentsContract.buildChildDocumentsUriUsingTree(tree, "bluray-output")
        return requireNotNull(resolver.query(children,
            arrayOf(DocumentsContract.Document.COLUMN_DISPLAY_NAME), null, null, null)).use { cursor ->
            buildList { while (cursor.moveToNext()) add(cursor.getString(0)) }
        }
    }

    private fun ensureMkv(
        outputUri: android.net.Uri, filename: String,
        wantVideo: Boolean, wantAudio: Boolean, chapters: Boolean,
    ) {
        check(DocumentsContract.getDocumentId(outputUri).substringAfterLast('/') == filename)
        val tracks = requireNotNull(resolver.openFileDescriptor(outputUri, "r")).use { fd ->
            check(fd.statSize > 0) { "Empty Blu-ray SAF output: $filename" }
            BlurayNativeIsoBridge(resolver).inspectMkv("/proc/self/fd/${fd.fd}")
        }
        check(tracks.mpeg2Video == (if (wantVideo) 1 else 0) &&
              tracks.flacAudio == (if (wantAudio) 1 else 0) && tracks.other == 0) {
            "Wrong Blu-ray output streams in $filename: $tracks"
        }
        check(tracks.chapters == (if (chapters) 2 else 0)) {
            "Blu-ray chapter selection lost for $filename: $tracks"
        }
        check(outputNames().none { it.contains(".partial-") || it == "source.mkv" }) {
            "Blu-ray created or leaked an intermediate MKV"
        }
        Log.i("muKsMaTTBlurayTest", "Device SAF MKV $filename: tracks=$tracks")
    }

    override fun onStart() {
        val result = Bundle()
        var status = Activity.RESULT_CANCELED
        try {
            resolver.call(providerUri, "prepare-bluray", null, null)
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
            ensureMkv(tree.remux(destination, "folder-full.mkv"),
                      "folder-full.mkv", true, true, true)
            ensureMkv(iso.remux(isoUri, destination, "iso-full.mkv"),
                      "iso-full.mkv", true, true, true)
            val audio = folderTracks.filter { it.type == "audio" }.map { it.index }.toIntArray()
            val video = folderTracks.filter { it.type == "video" }.map { it.index }.toIntArray()
            check(audio.isNotEmpty() && video.isNotEmpty())
            ensureMkv(
                tree.remux(destination, "audio-only.mkv", selectedStreamIndexes = audio,
                           includeChapters = false),
                "audio-only.mkv", false, true, false)
            ensureMkv(
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
            check(outputNames().toSet() == setOf("folder-full.mkv", "iso-full.mkv",
                "audio-only.mkv", "video-only.mkv")) { "Failed remux left unexpected output" }
            result.putString("stream", "MUKSMATT_BLURAY_SAF_PASS")
            status = Activity.RESULT_OK
        } catch (failure: Throwable) {
            Log.e("muKsMaTTBlurayTest", "Android Blu-ray device test failed", failure)
            result.putString("stream", failure.stackTraceToString())
        } finally {
            try {
                resolver.call(providerUri, "cleanup-bluray", null, null)
            } catch (failure: Throwable) {
                status = Activity.RESULT_CANCELED
                result.putString("stream", "SAF fixture cleanup failed: ${failure.stackTraceToString()}")
            }
        }
        finish(status, result)
    }
}
