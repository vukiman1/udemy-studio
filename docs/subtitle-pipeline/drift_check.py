"""Flag spans where a translated SRT looks shifted against the English one.

Cue lengths of a correct translation track the source closely; when lines drift by one,
the lengths correlate better with a neighbour cue than with their own.
"""

import sys
from pathlib import Path
from statistics import correlation

sys.path.insert(0, str(Path(__file__).parent))
from srt_tool import parse_srt  # noqa: E402

WINDOW = 30
STEP = 10
MIN_ADVANTAGE = 0.2


def window_correlation(en: list[int], vi: list[int], start: int, shift: int) -> float:
    pairs = [
        (en[i], vi[i + shift])
        for i in range(start, min(start + WINDOW, len(en)))
        if 0 <= i + shift < len(vi)
    ]
    if len(pairs) < WINDOW // 2:
        return 0.0
    xs, ys = zip(*pairs)
    if len(set(xs)) < 2 or len(set(ys)) < 2:
        return 0.0
    return correlation(xs, ys)


def suspicious_spans(en_path: Path, vi_path: Path) -> list[str]:
    en = [len(text) for _, _, text in parse_srt(en_path)]
    vi = [len(text) for _, _, text in parse_srt(vi_path)]
    spans = []
    for start in range(0, max(1, len(en) - WINDOW + 1), STEP):
        aligned = window_correlation(en, vi, start, 0)
        shifted = max(window_correlation(en, vi, start, s) for s in (-1, 1))
        if shifted - aligned > MIN_ADVANTAGE:
            spans.append(f"cues {start + 1}-{start + WINDOW}: aligned={aligned:.2f} shifted={shifted:.2f}")
    return spans


def main() -> None:
    root = Path(sys.argv[1])
    flagged = 0
    for en_path in sorted(root.rglob("*_en.srt")):
        vi_path = en_path.with_name(en_path.name.replace("_en.srt", "_vi.srt"))
        if not vi_path.exists():
            continue
        spans = suspicious_spans(en_path, vi_path)
        if spans:
            flagged += 1
            print(vi_path.relative_to(root))
            for span in spans:
                print("   ", span)
    print(f"{flagged} file(s) flagged")


if __name__ == "__main__":
    main()
