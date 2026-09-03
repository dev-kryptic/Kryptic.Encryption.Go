#!/usr/bin/env python3
"""Resolve the git-tag version for Kryptic.Encryption.Go.

Go consumers fetch this module from GitHub
(`go get github.com/dev-kryptic/Kryptic.Encryption.Go@vX.Y.Z`), so a release is a
`vX.Y.Z` tag. Patch auto-increments from the latest tag when the incoming
VERSION file keeps the same major and minor.

If this commit already changed major or minor, that version is tagged as-is.
The first release (no tags yet) also keeps the incoming version.
"""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

VERSION_FILE = Path("VERSION")
STABLE = re.compile(r"^(\d+)\.(\d+)\.(\d+)$")
TAG = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")


def parse(version: str) -> tuple[int, int, int] | None:
    match = STABLE.match(version.strip())
    if not match:
        return None
    return int(match.group(1)), int(match.group(2)), int(match.group(3))


def read_version_file() -> str:
    if not VERSION_FILE.exists():
        raise SystemExit(f"No {VERSION_FILE} found")
    value = VERSION_FILE.read_text(encoding="utf-8").strip()
    if not value:
        raise SystemExit(f"{VERSION_FILE} is empty")
    return value


def latest_published() -> tuple[int, int, int] | None:
    result = subprocess.run(
        ["git", "tag", "--list", "v*.*.*"],
        check=True,
        capture_output=True,
        text=True,
    )
    parsed: list[tuple[int, int, int]] = []
    for line in result.stdout.splitlines():
        match = TAG.match(line.strip())
        if match:
            parsed.append((int(match.group(1)), int(match.group(2)), int(match.group(3))))
    return max(parsed) if parsed else None


def render(version: tuple[int, int, int]) -> str:
    return f"{version[0]}.{version[1]}.{version[2]}"


def main() -> None:
    incoming_raw = sys.argv[1] if len(sys.argv) > 1 and sys.argv[1] else read_version_file()
    incoming = parse(incoming_raw)
    if incoming is None:
        raise SystemExit(f"Incoming version must be major.minor.patch, got: {incoming_raw}")

    published = latest_published()
    if published is None:
        resolved = incoming
        reason = "first release, keep incoming version"
    elif incoming[0] != published[0] or incoming[1] != published[1]:
        resolved = incoming
        reason = "major or minor changed, keep incoming version"
    else:
        resolved = (published[0], published[1], published[2] + 1)
        reason = "same major.minor, bump patch"

    print(f"incoming={render(incoming)}", file=sys.stderr)
    print(
        f"published={render(published) if published else '(none)'}",
        file=sys.stderr,
    )
    print(f"reason={reason}", file=sys.stderr)
    print(render(resolved))


if __name__ == "__main__":
    main()
