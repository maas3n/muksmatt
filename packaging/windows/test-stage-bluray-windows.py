"""Pure-stdlib packaging tests: no MinGW or installed Blu-ray libraries required."""
import importlib.util
from pathlib import Path
import tempfile
import types
import unittest
from unittest.mock import patch

MODULE = Path(__file__).with_name("stage-bluray-windows.py")
spec = importlib.util.spec_from_file_location("bluray_windows_stage", MODULE)
stage_mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(stage_mod)


class WindowsBlurayStagingTests(unittest.TestCase):
    def make_fixture(self, root):
        mingw = root / "mingw64" / "bin"
        mingw.mkdir(parents=True)
        for name in ("libbluray-4.dll", "libaacs-0.dll", "libbdplus-0.dll",
                     "libgcrypt-20.dll"):
            (mingw / name).write_bytes(name.encode())
        helper = root / "muksmatt-bluray-nav.exe"
        helper.write_bytes(b"test exe")
        objdump = root / "objdump.exe"
        objdump.write_bytes(b"dummy")
        for lib in ("libbluray", "libaacs", "libbdplus"):
            notice = root / "mingw64" / "share" / "licenses" / lib / "COPYING"
            notice.parent.mkdir(parents=True)
            notice.write_text("LGPL license text")
        return mingw, helper, objdump

    def fake_imports(self, binary):
        table = {
            "muksmatt-bluray-nav.exe": ["KERNEL32.dll", "libbluray-4.dll"],
            "libbluray-4.dll": ["WS2_32.dll"],
            "libaacs-0.dll": ["libgcrypt-20.dll", "api-ms-win-core-file-l1-1-0.dll"],
            "libbdplus-0.dll": ["libaacs-0.dll"],
            "libgcrypt-20.dll": ["msvcrt.dll"],
        }
        return table[binary.name]

    def fake_run(self, args, **kwargs):
        return types.SimpleNamespace(stdout="".join(
            "  DLL Name: " + dll + "\n" for dll in self.fake_imports(Path(args[2]))
        ))

    def test_recursively_stages_only_redistributable_dlls(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            mingw, helper, objdump = self.make_fixture(root)
            with patch.object(stage_mod.subprocess, "run", side_effect=self.fake_run):
                result = stage_mod.stage(helper, mingw, root / "out", objdump)
            self.assertEqual(set(result), {
                "muksmatt-bluray-nav.exe", "libbluray-4.dll",
                "libaacs-0.dll", "libbdplus-0.dll", "libgcrypt-20.dll"})
            self.assertFalse((root / "out" / "KERNEL32.dll").exists())
            self.assertTrue((root / "out" / "licenses" / "libbluray" / "COPYING").is_file())
            manifest = (root / "out" / "DEPENDENCIES.json").read_text()
            self.assertIn("libgcrypt-20.dll", manifest)
            self.assertNotIn("KEYDB.cfg", manifest)

    def test_missing_dependency_fails_closed(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            mingw, helper, objdump = self.make_fixture(root)
            with patch.object(stage_mod.subprocess, "run", side_effect=self.fake_run):
                (mingw / "libgcrypt-20.dll").unlink()
                with self.assertRaisesRegex(RuntimeError, "Unresolved MinGW DLL"):
                    stage_mod.stage(helper, mingw, root / "out", objdump)

    def test_refuses_to_stage_absent_protection_libraries(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            mingw, helper, objdump = self.make_fixture(root)
            (mingw / "libaacs-0.dll").unlink()
            with self.assertRaisesRegex(RuntimeError, "libaacs"):
                stage_mod.stage(helper, mingw, root / "out", objdump)

    def test_system_dll_filter(self):
        for name in ("Kernel32.DLL", "API-MS-WIN-CORE-PROCESSTHREADS-L1-1-0.dll",
                     "user32.dll", "WINMM.DLL"):
            self.assertTrue(stage_mod.is_system_dll(name))
        for name in ("libbluray-4.dll", "libgcc_s_seh-1.dll", "libaacs-0.dll"):
            self.assertFalse(stage_mod.is_system_dll(name))


if __name__ == "__main__":
    unittest.main()
