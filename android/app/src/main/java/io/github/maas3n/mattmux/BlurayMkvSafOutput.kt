package io.github.maas3n.mattmux

import android.content.ContentResolver
import android.net.Uri
import android.provider.DocumentsContract
import java.io.IOException

/**
 * Fail-closed Android SAF output transaction for a Blu-ray -> MKV remux.
 *
 * Never publish a truncated output under the final name. The source Blu-ray
 * transport is read directly by libbluray, not copied to a temporary MKV.
 */
internal object BlurayMkvSafOutput {
    fun create(
        resolver: ContentResolver,
        outputTreeUri: Uri,
        requestedName: String,
        write: (Int) -> Unit,
    ): Uri {
        require(DocumentsContract.isTreeUri(outputTreeUri)) {
            "Select an output folder with the system folder picker"
        }
        require(requestedName.length in 5..120 && requestedName.endsWith(".mkv", true) &&
            '/' !in requestedName && '\\' !in requestedName &&
            '\u0000' !in requestedName && requestedName != "..") {
            "Output must be a safe .mkv filename"
        }
        val parentId = DocumentsContract.getTreeDocumentId(outputTreeUri)
        val parent = DocumentsContract.buildDocumentUriUsingTree(outputTreeUri, parentId)
        val partial = DocumentsContract.createDocument(
            resolver, parent, "application/octet-stream",
            requestedName + ".partial-" + System.nanoTime().toString(36)
        ) ?: throw IOException("Could not create temporary Blu-ray MKV document")
        try {
            val descriptor = resolver.openFileDescriptor(partial, "w")
                ?: throw IOException("SAF provider cannot write MKV output")
            descriptor.use { write(it.fd) }
            return DocumentsContract.renameDocument(resolver, partial, requestedName)
                ?: throw IOException("Output provider cannot atomically publish Blu-ray MKV")
        } catch (error: Throwable) {
            runCatching { DocumentsContract.deleteDocument(resolver, partial) }
            throw error
        }
    }
}
