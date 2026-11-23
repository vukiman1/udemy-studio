"""Write per-file SRT digests from pysrt, the parser the Python downloader uses, as the oracle for Go's parser.

usage: pysrt_golden.py <library_dir> <out.json>
"""

import hashlib
import json
import sys
from pathlib import Path

import pysrt


def digest(lines: list[str]) -> str:
    return hashlib.sha256("\n".join(lines).encode("utf-8")).hexdigest()


def describe(path: Path) -> dict[str, object]:
    cues = pysrt.open(str(path), encoding="utf-8")
    return {
        "cue_count": len(cues),
        "timing_sha256": digest([f"{cue.index}\t{cue.start.ordinal}\t{cue.end.ordinal}" for cue in cues]),
        "text_sha256": digest([cue.text for cue in cues]),
    }


def main() -> None:
    library_dir, out_path = Path(sys.argv[1]), Path(sys.argv[2])
    golden = {path.relative_to(library_dir).as_posix(): describe(path) for path in sorted(library_dir.rglob("*.srt"))}
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(golden, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print(f"wrote {out_path} ({len(golden)} files)")


if __name__ == "__main__":
    main()
