"""Write VTT fixtures, the SRT the Python downloader converts each to, and Python's HTML unescape tables.

Usage: vtt_golden.py <fixture_dir> <html_entities.json>
Run with the udemy-downloader venv and PYTHONPATH set to that repo.
"""

import html
import html.entities
import json
import sys
from pathlib import Path

from udemy_downloader.vtt_to_srt import convert

FIXTURES = {
    "basic": (
        "WEBVTT\n\n"
        "1\n00:00:01.000 --> 00:00:04.000\nHello there.\n\n"
        "intro-2\n00:00:05.250 --> 00:01:02.500\nTwo lines\nof text.\n\n"
        "00:01:03.000 --> 00:01:04.000\nNo identifier.\n"
    ).encode(),
    "tags_and_entities": (
        "WEBVTT\n\n"
        "00:00:01.000 --> 00:00:02.000\n<v Roger Bingham>We are in <c.yellow>New York</c>.</v>\n\n"
        "00:00:02.000 --> 00:00:03.000\na < b <i>c</i> and <b>bold\n\n"
        "00:00:03.000 --> 00:00:04.000\n&amp; &lt;b&gt; &nbsp;x &#39; &#x27; &ampx &notit; &unknown; &\n\n"
        "00:00:04.000 --> 00:00:05.000\n&#0; [&#1;] [&#13;] &#128; &#xD800; &#99999999999; &#x1F600; &#xFDD0;x\n\n"
        "00:00:05.000 --> 00:00:06.000\nTiếng Việt &aacute; &Aacute &copy2024 &#;\n"
    ).encode(),
    "timestamps": (
        "WEBVTT\n\n"
        "0:1:5.2 --> 0:1:6.25\nShort fields.\n\n"
        "01:02.5 --> 01:03.000\nNo hours.\n\n"
        "123:00:00.000 --> 123:00:01.000\nThree digit hours.\n\n"
        "00:60:00.000 --> 00:61:00.000\nMinutes over 59 are dropped.\n\n"
        "12345678901:00:00.000 --> 12345678901:00:01.000\nEleven digit hours are dropped.\n\n"
        "00:00:01.000 --> 00:00:04.3001\nMalformed end keeps its milliseconds.\n\n"
        "  00:00:05.000-->00:00:06.000 align:start line:0%\nSettings and spacing.\n"
    ).encode(),
    "blocks": (
        "WEBVTT - with header text\nKind: captions\nLanguage: en\n\n"
        "NOTE a comment\nthat spans lines\n\n"
        "STYLE\n::cue { color: red }\n\n"
        "REGION\nid:fred width:40%\n\n"
        "00:00:01.000 --> 00:00:02.000\nfirst\n"
        "00:00:02.000 --> 00:00:03.000\nno blank line before me\n\n"
        "id with spaces\n00:00:03.000 --> 00:00:04.000\n   \nwhitespace-only line above\n\n"
        "00:00:04.000 --> 00:00:05.000\n\n"
        "not a cue\njust text\n\n"
        "00:00:06.000 --> 00:00:07.000\nline one\nline two\n00:00:07.000 --> 00:00:08.000\nsplit by an arrow\n"
    ).encode(),
    "bom_and_crlf": b"\xef\xbb\xbfWEBVTT\r\n\r\n00:00:01.000 --> 00:00:02.000\r\nCRLF line\r\n\r\n00:00:02.000 --> 00:00:03.000\rlone CR\r\rnext\n",
    "double_bom": "﻿﻿WEBVTT\n\n00:00:01.000 --> 00:00:02.000\ntwo byte order marks\n".encode(),
    "utf16le": b"\xff\xfe" + "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nUTF-16 LE ✓\n".encode("utf-16-le"),
    "utf16be": b"\xfe\xff" + "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nUTF-16 BE 😀\n".encode("utf-16-be"),
    "utf32le": b"\xff\xfe\x00\x00" + "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nUTF-32 LE\n".encode("utf-32-le"),
    "unicode_digits_and_spaces": (
        "WEBVTT\n\n"
        "٠٠:٠١.٥٠٠ -->　٠٠:٠٢.٠٠٠\nArabic-Indic digits.\n\n"
        " x\n00:00:03.000 --> 00:00:04.000\nLine separator stays in the identifier.\n"
    ).encode(),
    "arrows_without_blank_lines": (
        "WEBVTT\n00:00:01.000 --> 00:00:02.000\nright after the header\n\n"
        "00:00:02.000 --> 00:00:03.000\n00:00:03.000 --> 00:00:04.000\ntimings right after timings\n"
    ).encode(),
    "header_only": b"WEBVTT\n",
    "no_trailing_newline": b"WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nlast line",
    "missing_header": b"00:00:01.000 --> 00:00:02.000\nno WEBVTT line\n",
    "empty": b"",
    "invalid_utf8": b"WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nbad \xff byte\n",
}

RECORDED_SUFFIXES = (".vtt", ".srt", ".error")


def write_fixtures(fixture_dir: Path) -> None:
    fixture_dir.mkdir(parents=True, exist_ok=True)
    for name in FIXTURES:
        for suffix in RECORDED_SUFFIXES:
            (fixture_dir / f"{name}{suffix}").unlink(missing_ok=True)
    for name, content in FIXTURES.items():
        (fixture_dir / f"{name}.vtt").write_bytes(content)
        try:
            convert(str(fixture_dir), name)
        except Exception as error:  # the downloader records any failure as a failed caption
            (fixture_dir / f"{name}.srt").unlink(missing_ok=True)
            (fixture_dir / f"{name}.error").write_text(type(error).__name__ + "\n", encoding="utf-8")


def write_entities(path: Path) -> None:
    tables = {
        "html5": html.entities.html5,
        "invalid_charrefs": {str(code): text for code, text in html._invalid_charrefs.items()},
        "invalid_codepoints": sorted(html._invalid_codepoints),
    }
    path.write_text(json.dumps(tables, ensure_ascii=False, sort_keys=True), encoding="utf-8")


if __name__ == "__main__":
    write_fixtures(Path(sys.argv[1]))
    write_entities(Path(sys.argv[2]))
