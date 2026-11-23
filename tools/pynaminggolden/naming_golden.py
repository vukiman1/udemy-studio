"""Record how the Python downloader names files, as the oracle for internal/naming.

usage: naming_golden.py <course_content.json> <emoji_codes.json> <naming_golden.json>
"""

import json
import sys

import demoji
from pathvalidate import sanitize_filename

EDGE_CASES = [
    "HTTP/2 vs HTTP/3: why?",
    'a<b>c|d"e*f',
    "back\\slash and\ttab\nnewline\rreturn\x0bvt\x0cff\x00nul\x1fus\x7fdel",
    "name. ",
    "name...",
    "  lead",
    "   ",
    "",
    "\x1c",
    "a\x1cb",
    ".",
    "..",
    "...",
    "....",
    ".hidden",
    "  .hidden",
    "file.",
    " . ",
    "CON",
    "con",
    "con.txt",
    "CON.tar.gz",
    "aux.aux",
    "CON ",
    " CON",
    "NUL.",
    "COM1",
    "com9.log",
    "LPT¹",
    "COM³x",
    "CLOCK$",
    "clock$.txt",
    "PRN-1",
    "Kết nối TCP và độ trễ: Đầu tiên?",
    "Tiếng Việt có dấu – en dash — em dash",
    "emoji \U0001F680 test",
    "family \U0001F468‍\U0001F469‍\U0001F467 zwj",
    "flag \U0001F1FB\U0001F1F3 vn",
    "keycap 1️⃣ one",
    "thumbs \U0001F44D\U0001F3FD skin",
    "heart ❤️ red",
    "copyright © trademark ™ symbols",
    "only \U0001F600\U0001F601",
    "\U0001F600",
    "x" * 300,
    "é" * 200,
    "\U0001F680" * 70,
    "a" * 253 + "é",
    "a" * 254 + "é",
    "001 " + "Long lecture title " * 20 + ".mp4",
]


def titles_from_curriculum(path):
    with open(path, encoding="utf-8") as handle:
        content = json.load(handle)
    titles = []
    for item in content["results"]:
        titles.append(item.get("title") or "")
        asset = item.get("asset") or {}
        if asset.get("filename"):
            titles.append(asset["filename"])
        for supplementary in item.get("supplementary_assets") or []:
            titles.append(supplementary.get("title") or "")
            titles.append(supplementary.get("filename") or "")
    titles.append(content.get("course_title") or "")
    return titles


def main():
    curriculum_path, codes_path, golden_path = sys.argv[1:4]
    demoji.set_emoji_pattern()
    with open(codes_path, "w", encoding="utf-8") as handle:
        ordered = sorted(demoji._CODE_TO_DESC, key=len, reverse=True)
        json.dump(ordered, handle, ensure_ascii=False, indent=0)
        handle.write("\n")
    inputs = list(dict.fromkeys(titles_from_curriculum(curriculum_path) + EDGE_CASES))
    golden = [
        {"input": text, "sanitized": sanitize_filename(text), "deemojified": demoji.replace(text, "")}
        for text in inputs
    ]
    with open(golden_path, "w", encoding="utf-8") as handle:
        json.dump(golden, handle, ensure_ascii=False, indent=1)
        handle.write("\n")
    print(f"wrote {len(ordered)} emoji codes and {len(golden)} naming cases")


if __name__ == "__main__":
    main()
