package io.github.maas3n.mattmux

import android.content.Intent
import android.content.pm.PackageManager
import android.os.Binder
import android.os.Bundle
import android.os.Process
import android.provider.DocumentsContract
import android.database.Cursor
import android.database.MatrixCursor
import android.os.CancellationSignal
import android.os.ParcelFileDescriptor
import android.provider.DocumentsContract.Document
import android.provider.DocumentsProvider
import java.io.File

/** Debug-only real file descriptors behind SAF, for end-to-end DVD export tests. */
class DemuxDocumentsProvider : DocumentsProvider() {
    private val paths by lazy { DemuxDocumentPaths(File(context!!.cacheDir, "dvd-saf-test")) }
    private fun file(id: String): File = paths.file(id)
    override fun onCreate() = true

    // Fixtures live in the provider APK's own sandbox. A separate test APK has
    // a different UID from the signed application under instrumentation.
    override fun call(method: String, arg: String?, extras: Bundle?): Bundle? {
        val inherited = super.call(method, arg, extras)
        if (inherited != null || (method != "prepare-bluray" && method != "cleanup-bluray")) {
            return inherited
        }
        val ctx = requireNotNull(context)
        check(ctx.packageManager.checkSignatures(Binder.getCallingUid(), Process.myUid()) ==
            PackageManager.SIGNATURE_MATCH) { "Only the matching signed test application may prepare fixtures" }
        val root = File(ctx.cacheDir, "dvd-saf-test")
        check(root.deleteRecursively())
        if (method == "cleanup-bluray") return Bundle()
        check(root.mkdirs())
        fun unpack(source: String, destination: File) {
            val names = ctx.assets.list(source).orEmpty()
            if (names.isEmpty()) {
                destination.parentFile?.mkdirs()
                ctx.assets.open(source).use { input -> destination.outputStream().use { input.copyTo(it) } }
            } else {
                check(destination.mkdirs() || destination.isDirectory)
                names.forEach { unpack("$source/$it", File(destination, it)) }
            }
        }
        unpack("bluray-smoke/BDMV", File(root, "bluray-disc/BDMV"))
        if (!ctx.assets.list("bluray-smoke/CERTIFICATE").isNullOrEmpty()) {
            unpack("bluray-smoke/CERTIFICATE", File(root, "bluray-disc/CERTIFICATE"))
        }
        unpack("bluray-smoke/movie.iso", File(root, "movie.iso"))
        check(File(root, "bluray-disc/BDMV/PLAYLIST").isDirectory)
        check(File(root, "bluray-disc/BDMV/STREAM").isDirectory)
        check(File(root, "bluray-output").mkdirs())
        val authority = ctx.packageName + ".demux-test"
        val flags = Intent.FLAG_GRANT_READ_URI_PERMISSION or Intent.FLAG_GRANT_WRITE_URI_PERMISSION
        for (id in listOf("bluray-disc", "bluray-output")) {
            ctx.grantUriPermission("io.github.maas3n.muksmatt",
                DocumentsContract.buildTreeDocumentUri(authority, id),
                flags or Intent.FLAG_GRANT_PREFIX_URI_PERMISSION)
        }
        ctx.grantUriPermission("io.github.maas3n.muksmatt",
            DocumentsContract.buildDocumentUri(authority, "movie.iso"), flags)
        return Bundle()
    }
    override fun isChildDocument(parentDocumentId: String, documentId: String) = paths.isChild(parentDocumentId, documentId)
    override fun queryRoots(projection: Array<out String>?): Cursor = MatrixCursor(projection ?: emptyArray())
    private fun rows(projection: Array<out String>?, files: List<File>): Cursor {
        val columns = projection ?: arrayOf(Document.COLUMN_DOCUMENT_ID, Document.COLUMN_DISPLAY_NAME, Document.COLUMN_MIME_TYPE, Document.COLUMN_SIZE, Document.COLUMN_FLAGS)
        return MatrixCursor(columns).apply {
            for (f in files) addRow(columns.map { column -> when(column) {
                Document.COLUMN_DOCUMENT_ID -> paths.documentId(f)
                Document.COLUMN_DISPLAY_NAME -> f.name
                Document.COLUMN_MIME_TYPE -> if (f.isDirectory) Document.MIME_TYPE_DIR else "application/octet-stream"
                Document.COLUMN_SIZE -> f.length()
                Document.COLUMN_FLAGS -> Document.FLAG_SUPPORTS_WRITE or Document.FLAG_SUPPORTS_DELETE or (if (f.isDirectory) Document.FLAG_DIR_SUPPORTS_CREATE else 0)
                else -> null
            } }.toTypedArray())
        }
    }
    override fun queryDocument(documentId: String, projection: Array<out String>?) = rows(projection, listOf(file(documentId)))
    override fun queryChildDocuments(parentDocumentId: String, projection: Array<out String>?, sortOrder: String?) = rows(projection, file(parentDocumentId).listFiles()!!.toList())
    override fun openDocument(documentId: String, mode: String, signal: CancellationSignal?) = ParcelFileDescriptor.open(file(documentId), ParcelFileDescriptor.parseMode(mode))
    override fun createDocument(parentDocumentId: String, mimeType: String, displayName: String): String {
        val output = File(file(parentDocumentId), displayName)
        check(if (mimeType == Document.MIME_TYPE_DIR) output.mkdir() else output.createNewFile())
        return paths.documentId(output)
    }
    override fun renameDocument(documentId: String, displayName: String): String {
        require(displayName.isNotBlank() && displayName != "." && displayName != ".." &&
            '/' !in displayName && '\\' !in displayName) { "Unsafe document name" }
        val old = file(documentId)
        val parent = old.parentFile ?: error("Cannot rename root document")
        val newFile = File(parent, displayName)
        check(!newFile.exists() && old.renameTo(newFile)) { "Cannot publish SAF output" }
        return paths.documentId(newFile)
    }

    override fun deleteDocument(documentId: String) { check(file(documentId).deleteRecursively()) }
}
