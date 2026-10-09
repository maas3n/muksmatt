package io.github.maas3n.mattmux

import android.content.ContentResolver
import android.net.Uri
import android.provider.DocumentsContract
import android.provider.DocumentsContract.Document
import java.io.IOException
import java.util.ArrayDeque

/** Read-only SAF file and directory broker for libbluray bd_open_files(). */
internal class BluraySafTreeBridge(
    private val resolver: ContentResolver,
    private val treeUri: Uri,
    private val checkCancellation: () -> Unit = {},
) {
    private val index = BluraySafPathIndex()

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
    }

    internal data class Probe(val playlist: Int, val duration90kHz: Long, val bytesRead: Int)

    private external fun nativeInspectTree(
        provider: BluraySafTreeBridge,
        playlist: Int,
        sampleBytes: Int,
    ): String

    init {
        require(DocumentsContract.isTreeUri(treeUri)) { "Select a Blu-ray folder using the SAF picker" }
        buildIndex()
    }

    private fun documentUri(documentId: String): Uri =
        DocumentsContract.buildDocumentUriUsingTree(treeUri, documentId)

    private data class Entry(val id: String, val name: String, val directory: Boolean)

    private fun queryOne(id: String): Entry =
        resolver.query(
            documentUri(id),
            arrayOf(Document.COLUMN_DOCUMENT_ID, Document.COLUMN_DISPLAY_NAME, Document.COLUMN_MIME_TYPE),
            null, null, null,
        )?.use { c ->
            require(c.moveToFirst()) { "SAF document not found" }
            Entry(c.getString(0), c.getString(1) ?: "",
                c.getString(2) == Document.MIME_TYPE_DIR)
        } ?: throw IOException("SAF provider cannot inspect Blu-ray tree")

    private fun list(id: String): List<Entry> {
        val uri = DocumentsContract.buildChildDocumentsUriUsingTree(treeUri, id)
        return resolver.query(
            uri,
            arrayOf(Document.COLUMN_DOCUMENT_ID, Document.COLUMN_DISPLAY_NAME, Document.COLUMN_MIME_TYPE),
            null, null, null,
        )?.use { c ->
            val out = ArrayList<Entry>()
            while (c.moveToNext()) {
                checkCancellation()
                require(out.size < 10000) { "Excessive Blu-ray folder entries" }
                out.add(Entry(c.getString(0), c.getString(1) ?: "",
                    c.getString(2) == Document.MIME_TYPE_DIR))
            }
            out
        } ?: throw IOException("SAF provider cannot enumerate Blu-ray folder")
    }

    private fun buildIndex() {
        val rootId = DocumentsContract.getTreeDocumentId(treeUri)
        val root = queryOne(rootId)
        require(root.directory) { "Blu-ray source must be a directory" }
        index.register("", root.id, "", true)
        val pending = ArrayDeque<Pair<String, String>>()
        if (root.name.equals("BDMV", true)) {
            index.register("BDMV", root.id, "BDMV", true)
            pending.add("BDMV" to root.id)
        } else {
            pending.add("" to root.id)
        }
        var counted = 0
        while (pending.isNotEmpty()) {
            checkCancellation()
            val (prefix, parentId) = pending.removeFirst()
            for (entry in list(parentId)) {
                val path = if (prefix.isEmpty()) entry.name else prefix + "/" + entry.name
                index.register(path, entry.id, entry.name, entry.directory)
                counted++
                require(counted <= 30000) { "Blu-ray folder tree contains too many documents" }
                if (entry.directory) pending.add(path to entry.id)
            }
        }
        require(index.node("BDMV")?.isDirectory == true) { "BDMV directory not found" }
        require(index.node("BDMV/PLAYLIST")?.isDirectory == true) { "Blu-ray PLAYLIST not found" }
        require(index.node("BDMV/STREAM")?.isDirectory == true) { "Blu-ray STREAM not found" }
    }

    @JvmName("nativeListNames")
    fun nativeListNames(path: String): Array<String>? {
        checkCancellation()
        return index.names(path)
    }

    @JvmName("nativeOpenFd")
    fun nativeOpenFd(path: String): Int {
        checkCancellation()
        val entry = index.node(path) ?: return -1
        if (entry.isDirectory) return -1
        return resolver.openFileDescriptor(documentUri(entry.documentId), "r")?.use {
            // detachFd transfers ownership to the native BD_FILE_H close callback.
            it.detachFd()
        } ?: -1
    }

    fun inspect(playlist: Int? = null, sampleBytes: Int = 6144): Probe {
        require(playlist == null || playlist in 0..99999) { "Invalid playlist number" }
        require(sampleBytes in 0..6144) { "Probe must be bounded to 6144 bytes" }
        if (loadError != null) throw IOException("Native Blu-ray library unavailable", loadError)
        val parsed = BlurayNativeIsoBridge.parseNavigation(nativeInspectTree(this, playlist ?: -1, sampleBytes))
        return Probe(parsed.playlist, parsed.duration90kHz, parsed.bytesRead)
    }
}
