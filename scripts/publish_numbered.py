"""Stage, verify and only then publish one immutable numbered GitHub Release.

A failed or incomplete upload stays private as a draft. Unlike a streaming
"gh release create <tag> assets/*", this never exposes a partial public beta.
"""
import os
from pathlib import Path
import re
import sys

from publish_main import command, release_by_tag, verify_assets


def publish(root, repo, tag, version, is_prerelease, run=command):
    if not re.fullmatch(r"v\d+\.\d+\.\d+(?:-(?:alpha|beta|rc)\.\d+)?", tag):
        raise ValueError(f"Unsupported release tag: {tag}")
    if version != tag.removeprefix("v"):
        raise ValueError("Tag/version mismatch")
    if is_prerelease != ("-" in tag):
        raise ValueError("Tag and prerelease setting disagree")

    # The Git tag must be an already-created immutable reference to the exact
    # code currently checked out, not an accidental new tag or newer main.
    head = run("git", "rev-parse", "HEAD")
    tagged = run("git", "rev-parse", "--verify", f"refs/tags/{tag}^{{commit}}")
    if tagged != head:
        raise RuntimeError("Release tag does not refer to the checked-out commit")

    local = {path.name: path for path in root.iterdir() if path.is_file()}
    if not local or any(path.stat().st_size <= 0 for path in local.values()):
        raise RuntimeError("Empty or missing release payload")
    if "SHA256SUMS.txt" not in local:
        raise RuntimeError("Missing checksum manifest")

    notes_file = Path("docs/releases") / (tag + ".md")
    if is_prerelease and (not notes_file.is_file() or not notes_file.stat().st_size):
        raise RuntimeError(f"Missing reviewed prerelease notes: {notes_file}")

    # Always create a private draft FIRST and refuse to alter existing releases.
    # A rerun after an interrupted draft requires explicit inspection/cleanup.
    create = [
        "gh", "release", "create", tag, "--repo", repo, "--verify-tag",
        "--target", head, "--draft", "--latest=false",
        "--title", f"muKsMaTT {version}",
    ]
    if is_prerelease:
        create.append("--prerelease")
    if notes_file.is_file() and notes_file.stat().st_size:
        create.extend(("--notes-file", str(notes_file)))
    else:
        create.append("--generate-notes")
    run(*create)

    run("gh", "release", "upload", tag,
        *[str(local[name]) for name in sorted(local)], "--repo", repo)

    release = release_by_tag(repo, tag, run=run)
    if not release.get("draft") or release.get("tag_name") != tag:
        raise RuntimeError("Expected exactly the private draft created for this tag")
    if release.get("target_commitish") != head:
        raise RuntimeError("Draft targets a different commit")
    verify_assets(local, release.get("assets", []))

    # Only fully-uploaded and verified releases become visible to testers.
    edit = ["gh", "release", "edit", tag, "--repo", repo, "--draft=false",
            "--latest=false", f"--prerelease={str(is_prerelease).lower()}"]
    run(*edit)


if __name__ == "__main__":
    if len(sys.argv) != 5:
        raise SystemExit("Usage: publish_numbered.py ASSET_DIR TAG VERSION true|false")
    if sys.argv[4] not in ("true", "false"):
        raise SystemExit("Expected true or false for prerelease")
    publish(Path(sys.argv[1]), os.environ["GITHUB_REPOSITORY"],
            sys.argv[2], sys.argv[3], sys.argv[4] == "true")
