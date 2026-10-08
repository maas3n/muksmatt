#!/usr/bin/env python3
"""Bundle a Windows libbluray navigator and its MinGW runtime dependencies.

The helper imports libbluray as a DLL. libaacs/libbdplus are dynamically
loaded by libbluray, so include their DLLs as explicit dependency roots.
No keys, certificates, BD+ VM files or untrusted system DLLs are bundled.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess

# Native Windows libraries must be loaded from the operating system. Never
# copy them from MinGW directories into a private app folder.
WINDOWS_DLLS = {
    "advapi32.dll", "bcrypt.dll", "cabinet.dll", "comctl32.dll",
    "comdlg32.dll", "crypt32.dll", "d2d1.dll", "dwmapi.dll",
    "gdi32.dll", "imm32.dll", "iphlpapi.dll", "kernel32.dll",
    "msvcrt.dll", "ncrypt.dll", "netapi32.dll", "normaliz.dll",
    "ntdll.dll", "ole32.dll", "oleaut32.dll", "opengl32.dll",
    "psapi.dll", "rpcrt4.dll", "secur32.dll", "setupapi.dll",
    "shell32.dll", "shlwapi.dll", "ucrtbase.dll", "user32.dll",
    "userenv.dll", "version.dll", "winhttp.dll", "wininet.dll",
    "winmm.dll", "ws2_32.dll", "wsock32.dll",
}
# These are meant to be installed by MSYS2 only. No system MSYS2 runtime
# is included in an end-user Windows package.
OPTIONAL_ROOTS = ("libaacs-0.dll", "libbdplus-0.dll")
LICENSES = ("libbluray", "libaacs", "libbdplus")


def imported_dlls(binary: Path, objdump: Path) -> set[str]:
    result = subprocess.run([str(objdump), "-p", str(binary)], capture_output=True,
                            text=True, check=True)
    return set(re.findall(r"^\s*DLL Name:\s*(\S+)\s*$",
                          result.stdout, re.MULTILINE | re.IGNORECASE))


def is_system_dll(name: str) -> bool:
    lower = name.lower()
    return (lower in WINDOWS_DLLS or lower.startswith(("api-ms-win-", "ext-ms-win-"))
            or lower.endswith(".drv"))


def stage(helper: Path, mingw_bin: Path, dest: Path, objdump: Path) -> dict:
    helper = helper.resolve(strict=True)
    mingw_bin = mingw_bin.resolve(strict=True)
    objdump = objdump.resolve(strict=True)
    if not helper.is_file():
        raise ValueError("Blu-ray navigator executable missing")
    dll_map = {p.name.casefold(): p for p in mingw_bin.glob("*.dll")}
    for name in OPTIONAL_ROOTS + ("libbluray-4.dll",):
        if name.casefold() not in dll_map:
            raise RuntimeError(f"Required MinGW Blu-ray DLL not found: {name}")

    pending = [helper] + [dll_map[x.casefold()] for x in OPTIONAL_ROOTS]
    seen: set[str] = set()
    staged: dict[str, Path] = {}
    while pending:
        file = pending.pop()
        identity = file.name.casefold()
        if identity in seen:
            continue
        seen.add(identity)
        if file != helper:
            staged[file.name] = file
        for name in sorted(imported_dlls(file, objdump)):
            if is_system_dll(name):
                continue
            match = dll_map.get(name.casefold())
            if match is None:
                raise RuntimeError(f"Unresolved MinGW DLL import {name} from {file.name}")
            pending.append(match)
    if "libbluray-4.dll" not in {x.casefold() for x in staged}:
        raise RuntimeError("Navigator does not dynamically link libbluray-4.dll")
    dest.mkdir(parents=True, exist_ok=True)
    files = {"muksmatt-bluray-nav.exe": helper, **staged}
    manifest = {}
    for name, file in sorted(files.items()):
        target = dest / name
        shutil.copy2(file, target)
        manifest[name] = {
            "sha256": hashlib.sha256(target.read_bytes()).hexdigest(),
            "bytes": target.stat().st_size,
        }

    # Preserve redistributable project license texts; no example KEYDB.cfg.
    mingw_root = mingw_bin.parent
    for library in LICENSES:
        candidates = (
            mingw_root / "share" / "licenses" / library / "COPYING",
            mingw_root / "share" / "doc" / library / "COPYING",
        )
        notice = next((p for p in candidates if p.is_file()), None)
        if not notice:
            raise RuntimeError(f"Missing license text for {library}")
        target = dest / "licenses" / library / "COPYING"
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(notice, target)
    (dest / "DEPENDENCIES.json").write_text(
        json.dumps({"generator": "stage-bluray-windows.py",
                    "files": manifest, "license_roots": list(LICENSES)},
                   sort_keys=True, indent=2) + "\n",
        encoding="utf-8",
    )
    return manifest


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--helper", type=Path, required=True)
    parser.add_argument("--mingw-bin", type=Path, required=True)
    parser.add_argument("--dest", type=Path, required=True)
    parser.add_argument("--objdump", type=Path, required=True)
    args = parser.parse_args()
    manifest = stage(args.helper, args.mingw_bin, args.dest, args.objdump)
    print(f"Staged Blu-ray navigator and {len(manifest)-1} DLLs")
    for name in manifest:
        print(f"  {name}")


if __name__ == "__main__":
    main()
