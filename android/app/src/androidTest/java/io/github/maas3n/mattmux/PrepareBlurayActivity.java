package io.github.maas3n.mattmux;

import android.app.Activity;
import android.net.Uri;
import android.os.Bundle;

/** Bootstrap in the companion UID before instrumenting the production app. */
public final class PrepareBlurayActivity extends Activity {
    @Override public void onCreate(Bundle state) {
        super.onCreate(state);
        getContentResolver().call(Uri.parse("content://" + getPackageName() + ".demux-test"),
                "prepare-bluray", null, null);
        finish();
    }
}
