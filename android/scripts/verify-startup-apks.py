#!/usr/bin/env python3
"""Fail before emulator startup if release/test packaging or signing is wrong."""
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import zipfile


def only_apk(directory):
    paths = list(Path(directory).glob('*.apk'))
    if len(paths) != 1:
        raise RuntimeError(f'Expected exactly one APK in {directory}, found {len(paths)}')
    return paths[0]


def command(*args):
    return subprocess.check_output([str(a) for a in args], text=True, stderr=subprocess.STDOUT)


def verify(app, test, tools):
    manifests = [command(tools / 'aapt', 'dump', 'xmltree', apk, 'AndroidManifest.xml')
                 for apk in (app, test)]
    if 'E: instrumentation' in manifests[0] or 'DocumentsProvider' in manifests[0]:
        raise RuntimeError('CI instrumentation/provider leaked into the production APK')
    if 'BluraySmokeInstrumentation' not in manifests[1] or 'ReleaseDocumentsProvider' not in manifests[1]:
        raise RuntimeError('Test APK is missing required instrumentation/provider')
    if 'io.github.maas3n.muksmatt.test' not in manifests[1]:
        raise RuntimeError('Wrong test APK identity')
    if not re.search(r'android:targetPackage[^\n]*="io.github.maas3n.muksmatt"', manifests[1]):
        raise RuntimeError('Instrumentation must target the production app')
    certificates = []
    for apk in (app, test):
        report = command(tools / 'apksigner', 'verify', '--print-certs', apk)
        match = re.search(r'Signer #1 certificate SHA-256 digest: ([0-9a-fA-F]+)', report)
        if not match:
            raise RuntimeError(f'Cannot verify signing certificate for {apk}')
        certificates.append(match[1].lower())
    if certificates[0] != certificates[1]:
        raise RuntimeError('App and instrumentation APK signing identities differ')
    with zipfile.ZipFile(app) as archive:
        if any(name.startswith(('assets/bluray-smoke/', 'assets/demux-smoke/')) for name in archive.namelist()):
            raise RuntimeError('CI media fixtures leaked into the production APK')
    with zipfile.ZipFile(test) as archive:
        names = archive.namelist()
        # A manifest entry alone does not prove the Kotlin runner was compiled.
        descriptors = set()
        with tempfile.TemporaryDirectory() as directory:
            for name in names:
                if re.fullmatch(r'classes(?:\d+)?\.dex', name):
                    dex = Path(directory) / name
                    dex.write_bytes(archive.read(name))
                    report = command(tools / 'dexdump', dex)
                    descriptors.update(re.findall(r"Class descriptor\s*:\s*'([^']+)'", report))
        if 'Lio/github/maas3n/mattmux/BluraySmokeInstrumentation;' not in descriptors:
            raise RuntimeError('Test APK declares instrumentation but its runner class was not compiled')
        if ('assets/bluray-smoke/movie.iso' not in names
                or not any(n.startswith('assets/bluray-smoke/BDMV/PLAYLIST/') for n in names)
                or not any(n.startswith('assets/bluray-smoke/BDMV/STREAM/') for n in names)):
            raise RuntimeError('Test APK is missing authored ISO/BDMV fixtures')
    print('Production APK is isolated; separate instrumentation and fixtures are present; signing certificates match.')


if __name__ == '__main__':
    verify(only_apk(sys.argv[1]), only_apk(sys.argv[2]), Path(sys.argv[3]))
