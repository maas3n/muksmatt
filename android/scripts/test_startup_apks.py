"""Release APK checks must reject a manifest-only instrumentation runner."""
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import zipfile

spec = importlib.util.spec_from_file_location('verify_startup_apks', Path(__file__).with_name('verify-startup-apks.py'))
verifier = importlib.util.module_from_spec(spec)
spec.loader.exec_module(verifier)


class ApkVerificationTests(unittest.TestCase):
    def test_runner_must_be_defined_in_dex_and_certificates_must_match(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            app, test = root / 'app.apk', root / 'test.apk'
            with zipfile.ZipFile(app, 'w'):
                pass
            with zipfile.ZipFile(test, 'w') as archive:
                for name in ('classes.dex', 'assets/bluray-smoke/movie.iso',
                             'assets/bluray-smoke/BDMV/PLAYLIST/00000.mpls',
                             'assets/bluray-smoke/BDMV/STREAM/00000.m2ts'):
                    archive.writestr(name, b'fixture')
            dex_report = ''
            test_certificate = 'abcd'

            def command(*args):
                if args[0].name == 'aapt':
                    if args[3] == app:
                        return 'production manifest'
                    return ('BluraySmokeInstrumentation ReleaseDocumentsProvider '
                            'io.github.maas3n.muksmatt.test\n'
                            'android:targetPackage="io.github.maas3n.muksmatt"')
                if args[0].name == 'apksigner':
                    certificate = 'abcd' if args[-1] == app else test_certificate
                    return 'Signer #1 certificate SHA-256 digest: ' + certificate
                return dex_report

            with patch.object(verifier, 'command', side_effect=command):
                with self.assertRaisesRegex(RuntimeError, 'runner class was not compiled'):
                    verifier.verify(app, test, root)
                dex_report = "Class descriptor  : 'Lio/github/maas3n/mattmux/BluraySmokeInstrumentation;'"
                verifier.verify(app, test, root)
                test_certificate = 'ffff'
                with self.assertRaisesRegex(RuntimeError, 'signing identities differ'):
                    verifier.verify(app, test, root)


if __name__ == '__main__':
    unittest.main()
