"""Extract SRT cue text for translation, then merge translated text back onto the original timings.

  extract <en.srt> <out.txt>             one line per cue: "<index>\t<text>"
  merge   <en.srt> <out.srt> <vi.txt>... builds out.srt from en timings + translated lines, fails on any mismatch
  check   <root_dir>                     verifies every *_en.srt has a matching *_vi.srt
"""

import re
import sys
from pathlib import Path

TRANSLATED_LINE = re.compile(r"^(\d+)\t(.*)$")


class SrtMismatchError(Exception):
    pass


def parse_srt(path: Path) -> list[tuple[int, str, str]]:
    blocks = re.split(r"\n\s*\n", path.read_text(encoding="utf-8-sig").replace("\r\n", "\n").strip())
    cues = []
    for block in blocks:
        lines = block.split("\n")
        if len(lines) < 2 or "-->" not in lines[1]:
            raise SrtMismatchError(f"{path}: malformed block starting with {lines[0]!r}")
        cues.append((int(lines[0].strip()), lines[1].strip(), " ".join(line.strip() for line in lines[2:])))
    return cues


def read_translations(paths: list[Path]) -> dict[int, str]:
    translations: dict[int, str] = {}
    for path in paths:
        for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            if not line.strip():
                continue
            match = TRANSLATED_LINE.match(line)
            if not match:
                raise SrtMismatchError(f"{path}:{number}: expected '<index>\\t<text>', got {line[:80]!r}")
            index = int(match.group(1))
            if index in translations:
                raise SrtMismatchError(f"{path}:{number}: cue {index} translated twice")
            translations[index] = match.group(2).strip()
    return translations


def extract(en_path: Path, out_path: Path) -> None:
    cues = parse_srt(en_path)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text("".join(f"{index}\t{text}\n" for index, _, text in cues), encoding="utf-8")
    print(f"{len(cues)} cues -> {out_path}")


def merge(en_path: Path, out_path: Path, vi_paths: list[Path]) -> None:
    cues = parse_srt(en_path)
    translations = read_translations(vi_paths)
    expected = {index for index, _, _ in cues}
    missing = sorted(expected - translations.keys())
    extra = sorted(translations.keys() - expected)
    empty = sorted(index for index in expected & translations.keys() if not translations[index])
    if missing or extra or empty:
        raise SrtMismatchError(f"missing={missing[:20]} extra={extra[:20]} empty={empty[:20]}")
    body = "\n".join(f"{index}\n{timing}\n{translations[index]}\n" for index, timing, _ in cues)
    out_path.write_text(body, encoding="utf-8")
    print(f"OK {len(cues)} cues -> {out_path}")


def check(root: Path) -> None:
    problems = 0
    for en_path in sorted(root.rglob("*_en.srt")):
        vi_path = en_path.with_name(en_path.name.replace("_en.srt", "_vi.srt"))
        if not vi_path.exists():
            print(f"MISSING {vi_path.relative_to(root)}")
            problems += 1
            continue
        en_cues, vi_cues = parse_srt(en_path), parse_srt(vi_path)
        if [(i, t) for i, t, _ in en_cues] != [(i, t) for i, t, _ in vi_cues]:
            print(f"TIMING MISMATCH {vi_path.relative_to(root)}")
            problems += 1
    print(f"{problems} problem(s)")
    sys.exit(1 if problems else 0)


def main() -> None:
    command, *args = sys.argv[1:]
    try:
        if command == "extract":
            extract(Path(args[0]), Path(args[1]))
        elif command == "merge":
            merge(Path(args[0]), Path(args[1]), [Path(p) for p in args[2:]])
        elif command == "check":
            check(Path(args[0]))
        else:
            raise SystemExit(f"unknown command {command}")
    except SrtMismatchError as error:
        raise SystemExit(f"ERROR: {error}")


if __name__ == "__main__":
    main()
