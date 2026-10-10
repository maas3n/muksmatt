package io.github.maas3n.mattmux

import android.app.Activity
import android.net.Uri
import android.os.Bundle

/** CI-only bootstrap: seed provider-owned fixtures and grant SAF access before
 * instrumentation starts inside the distinct production application's UID. */
class PrepareBlurayActivity : Activity() {
    override fun onCreate(state: Bundle?) {
        super.onCreate(state)
        contentResolver.call(Uri.parse("content://$packageName.demux-test"),
            "prepare-bluray", null, null)
        finish()
    }
}
