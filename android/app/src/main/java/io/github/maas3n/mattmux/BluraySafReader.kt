package io.github.maas3n.mattmux

import android.content.ContentResolver
import android.net.Uri
import android.provider.DocumentsContract
import android.provider.DocumentsContract.Document
import java.io.ByteArrayOutputStream

/**
 * Read-only Blu-ray MPLS scanner for an explicitly granted SAF document tree.
 *
 * This is real provider I/O, not direct /storage access and not a fake path
 * handed to libbluray. Native decrypted M2TS reads require a later seekable
 * SAF bridge; this adapter exposes navigation metadata only.
 */
internal class BluraySafReader(
    private val resolver: ContentResolver,
    private val cancelled: () -> Unit = {},
) {
    private val maxPlaylistBytes = 16 * 1024 * 1024

    fun scan(treeUri: Uri): List<BluraySafPlaylist> {
        require(DocumentsContract.isTreeUri(treeUri)) {
            "Grant a Blu-ray folder using the system folder picker"
        }
        val rootId = DocumentsContract.getTreeDocumentId(treeUri)
        val rootUri = DocumentsContract.buildDocumentUriUsingTree(treeUri, rootId)
        cancelled()
        val root = resolver.query(
            rootUri,
            arrayOf(Document.COLUMN_DOCUMENT_ID, Document.COLUMN_DISPLAY_NAME, Document.COLUMN_MIME_TYPE),
            null, null, null
        )?.use { cursor ->
            require(cursor.moveToFirst()) { "Blu-ray folder is unavailable" }
            BlurayDocumentEntry(
                cursor.getString(0),
                cursor.getString(1) ?: "",
                cursor.getString(2) == Document.MIME_TYPE_DIR,
            )
        } ?: throw IllegalArgumentException("Blu-ray SAF provider did not return the folder")

        val provider = object : BlurayDocumentProvider {
            override fun listChildren(parentDocumentId: String): List<BlurayDocumentEntry> {
                cancelled()
                val uri = DocumentsContract.buildChildDocumentsUriUsingTree(treeUri, parentDocumentId)
                return resolver.query(
                    uri,
                    arrayOf(Document.COLUMN_DOCUMENT_ID, Document.COLUMN_DISPLAY_NAME, Document.COLUMN_MIME_TYPE),
                    null, null, null,
                )?.use { cursor ->
                    val entries = mutableListOf<BlurayDocumentEntry>()
                    while (cursor.moveToNext()) {
                        cancelled()
                        require(entries.size < 10_000) { "Blu-ray folder has too many documents" }
                        entries += BlurayDocumentEntry(
                            cursor.getString(0),
                            cursor.getString(1) ?: "",
                            cursor.getString(2) == Document.MIME_TYPE_DIR,
                        )
                    }
                    entries
                } ?: throw IllegalArgumentException("Blu-ray SAF provider could not list " + parentDocumentId)
            }

            override fun readPlaylist(documentId: String): ByteArray {
                cancelled()
                val uri = DocumentsContract.buildDocumentUriUsingTree(treeUri, documentId)
                val input = resolver.openInputStream(uri)
                    ?: throw IllegalArgumentException("Could not open Blu-ray playlist " + documentId)
                return input.use { stream ->
                    val buffer = ByteArray(32 * 1024)
                    ByteArrayOutputStream().use { result ->
                        while (true) {
                            cancelled()
                            val count = stream.read(buffer)
                            if (count < 0) break
                            require(result.size() + count <= maxPlaylistBytes) { "Blu-ray MPLS exceeds 16 MiB" }
                            result.write(buffer, 0, count)
                        }
                        result.toByteArray()
                    }
                }
            }
        }
        return BluraySafPlaylistCatalog.scan(root, provider)
    }
}
