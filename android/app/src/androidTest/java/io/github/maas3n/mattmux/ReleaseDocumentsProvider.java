package io.github.maas3n.mattmux;

import android.content.Context;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.database.Cursor;
import android.database.MatrixCursor;
import android.os.Binder;
import android.os.Bundle;
import android.os.CancellationSignal;
import android.os.ParcelFileDescriptor;
import android.os.Process;
import android.provider.DocumentsContract;
import android.provider.DocumentsContract.Document;
import android.provider.DocumentsProvider;
import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;

/** Test APK provider runs independently: use framework/Java only, since the
 * target APK's Kotlin runtime is available only in the instrumented process. */
public final class ReleaseDocumentsProvider extends DocumentsProvider {
    private File root;
    @Override public boolean onCreate() {
        try {
            root = new File(getContext().getCacheDir(), "dvd-saf-test").getCanonicalFile();
            android.util.Log.i("muKsMaTTBlurayTest", "Companion provider uid=" + Process.myUid());
            return true;
        } catch (IOException e) { throw new IllegalStateException(e); }
    }
    private File file(String id) throws FileNotFoundException {
        try {
            File result = new File(root, id).getCanonicalFile();
            if (!result.equals(root) && !result.getPath().startsWith(root.getPath() + File.separator))
                throw new IOException("Document escapes test root");
            return result;
        } catch (IOException e) { throw new FileNotFoundException(e.toString()); }
    }
    private String id(File value) throws IOException {
        return root.toPath().relativize(value.getCanonicalFile().toPath()).toString();
    }
    private static void remove(File value) throws IOException {
        File[] children = value.listFiles();
        if (children != null) for (File child : children) remove(child);
        if (value.exists() && !value.delete()) throw new IOException("Cannot remove " + value);
    }
    private void unpack(String asset, File destination) throws IOException {
        String[] children = getContext().getAssets().list(asset);
        if (children != null && children.length > 0) {
            if (!destination.mkdirs() && !destination.isDirectory()) throw new IOException("Cannot create fixture directory");
            for (String child : children) unpack(asset + "/" + child, new File(destination, child));
        } else {
            File parent = destination.getParentFile();
            if (!parent.mkdirs() && !parent.isDirectory()) throw new IOException("Cannot create fixture parent");
            try (InputStream input = getContext().getAssets().open(asset);
                 FileOutputStream output = new FileOutputStream(destination)) {
                byte[] buffer = new byte[65536];
                int count;
                while ((count = input.read(buffer)) != -1) output.write(buffer, 0, count);
            }
        }
    }
    @Override public Bundle call(String method, String arg, Bundle extras) {
        Bundle inherited = super.call(method, arg, extras);
        if (inherited != null || (!method.equals("prepare-bluray") && !method.equals("cleanup-bluray")))
            return inherited;
        Context ctx = getContext();
        if (ctx.getPackageManager().checkSignatures(Binder.getCallingUid(), Process.myUid()) != PackageManager.SIGNATURE_MATCH)
            throw new SecurityException("Fixture caller must have the same signing certificate");
        long identity = Binder.clearCallingIdentity();
        try {
            remove(root);
            if (method.equals("cleanup-bluray")) return new Bundle();
            if (!root.mkdirs()) throw new IOException("Cannot create fixture root");
            unpack("bluray-smoke/BDMV", new File(root, "bluray-disc/BDMV"));
            String[] certificate = ctx.getAssets().list("bluray-smoke/CERTIFICATE");
            if (certificate != null && certificate.length > 0)
                unpack("bluray-smoke/CERTIFICATE", new File(root, "bluray-disc/CERTIFICATE"));
            unpack("bluray-smoke/movie.iso", new File(root, "movie.iso"));
            if (!new File(root, "bluray-output").mkdirs()) throw new IOException("Cannot create output directory");
            String authority = ctx.getPackageName() + ".demux-test";
            int flags = Intent.FLAG_GRANT_READ_URI_PERMISSION | Intent.FLAG_GRANT_WRITE_URI_PERMISSION;
            for (String directory : new String[]{"bluray-disc", "bluray-output"})
                ctx.grantUriPermission("io.github.maas3n.muksmatt",
                    DocumentsContract.buildTreeDocumentUri(authority, directory), flags | Intent.FLAG_GRANT_PREFIX_URI_PERMISSION);
            ctx.grantUriPermission("io.github.maas3n.muksmatt",
                DocumentsContract.buildDocumentUri(authority, "movie.iso"), flags);
            return new Bundle();
        } catch (IOException e) { throw new IllegalStateException(e); }
        finally { Binder.restoreCallingIdentity(identity); }
    }
    @Override public boolean isChildDocument(String parent, String child) {
        try { return file(child).getPath().startsWith(file(parent).getPath() + File.separator); }
        catch (FileNotFoundException e) { return false; }
    }
    @Override public Cursor queryRoots(String[] projection) { return new MatrixCursor(projection == null ? new String[0] : projection); }
    private Cursor rows(String[] projection, File[] files) throws FileNotFoundException {
        String[] columns = projection == null ? new String[]{Document.COLUMN_DOCUMENT_ID,
            Document.COLUMN_DISPLAY_NAME, Document.COLUMN_MIME_TYPE, Document.COLUMN_SIZE, Document.COLUMN_FLAGS} : projection;
        MatrixCursor cursor = new MatrixCursor(columns);
        if (files == null) throw new FileNotFoundException("Cannot list test documents");
        for (File value : files) {
            MatrixCursor.RowBuilder row = cursor.newRow();
            for (String column : columns) {
                switch (column) {
                    case Document.COLUMN_DOCUMENT_ID:
                        try { row.add(id(value)); } catch (IOException e) { throw new FileNotFoundException(e.toString()); }
                        break;
                    case Document.COLUMN_DISPLAY_NAME: row.add(value.getName()); break;
                    case Document.COLUMN_MIME_TYPE: row.add(value.isDirectory() ? Document.MIME_TYPE_DIR : "application/octet-stream"); break;
                    case Document.COLUMN_SIZE: row.add(value.length()); break;
                    case Document.COLUMN_FLAGS: row.add(Document.FLAG_SUPPORTS_WRITE | Document.FLAG_SUPPORTS_DELETE |
                        Document.FLAG_SUPPORTS_RENAME | (value.isDirectory() ? Document.FLAG_DIR_SUPPORTS_CREATE : 0)); break;
                    default: row.add(null);
                }
            }
        }
        return cursor;
    }
    @Override public Cursor queryDocument(String document, String[] projection) throws FileNotFoundException {
        return rows(projection, new File[]{file(document)});
    }
    @Override public Cursor queryChildDocuments(String parent, String[] projection, String sort) throws FileNotFoundException {
        return rows(projection, file(parent).listFiles());
    }
    @Override public ParcelFileDescriptor openDocument(String document, String mode, CancellationSignal signal) throws FileNotFoundException {
        return ParcelFileDescriptor.open(file(document), ParcelFileDescriptor.parseMode(mode));
    }
    private static void checkName(String name) {
        if (name.isEmpty() || name.equals(".") || name.equals("..") || name.contains("/") || name.contains("\\"))
            throw new IllegalArgumentException("Unsafe document name");
    }
    @Override public String createDocument(String parent, String mime, String name) throws FileNotFoundException {
        checkName(name);
        File output = new File(file(parent), name);
        try {
            if (!(mime.equals(Document.MIME_TYPE_DIR) ? output.mkdir() : output.createNewFile())) throw new IOException("Output exists");
            return id(output);
        } catch (IOException e) { throw new FileNotFoundException(e.toString()); }
    }
    @Override public String renameDocument(String document, String name) throws FileNotFoundException {
        checkName(name);
        File old = file(document);
        if (old.equals(root)) throw new IllegalArgumentException("Cannot rename root");
        File output = new File(old.getParentFile(), name);
        if (output.exists() || !old.renameTo(output)) throw new FileNotFoundException("Cannot publish output");
        try { return id(output); } catch (IOException e) { throw new FileNotFoundException(e.toString()); }
    }
    @Override public void deleteDocument(String document) throws FileNotFoundException {
        try { remove(file(document)); } catch (IOException e) { throw new FileNotFoundException(e.toString()); }
    }
}
