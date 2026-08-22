#!/usr/bin/env python3
"""Fail CI when source files contain invisible Unicode used by GlassWorm."""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path


# GlassWorm encodes payloads in supplementary-plane Unicode variation
# selectors. The additional zero-width characters are commonly used to conceal
# or join a hidden payload. U+FE00–U+FE0F are deliberately not individually
# blocked because they are valid emoji presentation selectors; a large count is
# still suspicious and handled separately below.
SUSPICIOUS_CODEPOINTS = {
    *range(0xE0100, 0xE01F0),
    0x200B,
    0x200C,
    0x200D,
    0x2060,
    0xFEFF,
}


def tracked_files() -> list[Path]:
    result = subprocess.run(
        ["git", "ls-files", "-z"], check=True, capture_output=True
    )
    return [Path(name) for name in result.stdout.decode().split("\0") if name]


def main() -> int:
    findings: list[tuple[Path, int, int]] = []
    for path in tracked_files():
        try:
            content = path.read_text(encoding="utf-8")
        except (UnicodeDecodeError, OSError):
            continue

        emoji_selector_count = sum(0xFE00 <= ord(char) <= 0xFE0F for char in content)
        if emoji_selector_count >= 16:
            findings.append((path, 1, 0xFE00))

        for line_number, line in enumerate(content.splitlines(), start=1):
            for character in line:
                if ord(character) in SUSPICIOUS_CODEPOINTS:
                    findings.append((path, line_number, ord(character)))

    if not findings:
        print("GlassWorm indicator scan passed: no suspicious invisible Unicode found.")
        return 0

    for path, line_number, codepoint in findings:
        print(
            f"::error file={path},line={line_number}::"
            f"Suspicious invisible Unicode U+{codepoint:04X} detected"
        )
    return 1


if __name__ == "__main__":
    sys.exit(main())
