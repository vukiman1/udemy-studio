"""Record what the Python downloader would fetch and where it would write it, as the oracle for internal/udemy.

usage: plan_golden.py <course_content.json> <hls_dir> <plan_golden.json> <hls_golden.json>
Run with the downloader's venv python and its repo on PYTHONPATH.
"""

import copy
import json
import logging
import os
import re
import sys
import tempfile
from pathlib import Path
from urllib.parse import urlparse

from pathvalidate import sanitize_filename

from udemy_downloader.Udemy import Udemy
from udemy_downloader.utils import deEmojify

QUALITIES = [None, 144, 240, 360, 432, 480, 576, 600, 720, 900, 1080, 1440, 2160]
QUIZ_LIKE_CLASSES = ("quiz", "practice", "role-play")
CHILD_PLAYLIST = "#EXTM3U\n"
MASTER_ASSET_ID = re.compile(r"assets/(\d+)/")


class FixtureResponse:
    def __init__(self, text):
        self.text = text

    def raise_for_status(self):
        pass


class FixtureSession:
    """Serves each lecture's HLS master from the fixture dir and an empty playlist for every variant."""

    def __init__(self, hls_dir, content):
        self.variant_urls = []
        self.masters = {}
        for item in content["results"]:
            for source in (((item.get("asset") or {}).get("stream_urls") or {}).get("Video") or []):
                asset_id = MASTER_ASSET_ID.search(source.get("file") or "")
                if asset_id:
                    self.masters[source["file"]] = Path(hls_dir, f"index_{asset_id.group(1)}.m3u8")

    def _get(self, url, params=None):
        master = self.masters.get(url)
        if master is None:
            self.variant_urls.append(url)
            return FixtureResponse(CHILD_PLAYLIST)
        return FixtureResponse(master.read_text(encoding="utf-8"))

    def terminate(self):
        pass


def build_udemy(content, hls_dir, continuous_numbers, skip_hls):
    udemy = Udemy.__new__(Udemy)
    udemy.logger = logging.getLogger("plan-golden")
    udemy.session = FixtureSession(hls_dir, content)
    udemy.skip_hls = skip_hls
    udemy.use_continuous_lecture_numbers = continuous_numbers
    udemy.save_to_file = False
    udemy.info = True
    udemy.print_course_info = lambda: None
    udemy._extract_mpd = lambda url: []
    udemy.bearer_token = None
    udemy.course_id = content["course_id"]
    udemy.course_title = content["course_title"]
    udemy.course_slug = content["course_slug"]
    udemy.portal_name = content["portal_name"]
    udemy.course_data = copy.deepcopy(content)
    udemy.parsed_data = {}
    udemy.parse_course_data()
    return udemy


def video_choice(source):
    choice = {"type": source["type"], "height": int(source["height"]), "width": int(source["width"]), "url": source["download_url"]}
    if source["type"] == "hls":
        choice["url"] = source["variant_url"]
        choice["playlist_file"] = Path(urlparse(source["download_url"]).path).name
    return choice


def attach_variant_urls(sources, variant_urls):
    """_extract_m3u8 fetches each kept variant right before appending it, so fetch order is source order."""
    hls_sources = [source for source in sources if source["type"] == "hls"]
    for source, url in zip(hls_sources, variant_urls, strict=True):
        source["variant_url"] = url


def choose(sources, quality):
    ranked = sorted(sources, key=lambda x: int(x.get("height")), reverse=True)
    if not ranked:
        return None
    if quality is None:
        return video_choice(ranked[0])
    return video_choice(min(ranked, key=lambda x: abs(int(x.get("height")) - quality)))


def caption_plan(lecture_title, caption):
    stem = "%s_%s" % (sanitize_filename(lecture_title), caption["language"])
    return {
        "language": caption["language"],
        "url": caption["download_url"],
        "file": "%s.%s" % (stem, caption["extension"]),
        "srt_file": f"{stem}.srt",
    }


def asset_plan(lecture_title, asset):
    kind = asset["type"]
    if kind == "article":
        return {"type": kind, "file": "{}.html".format(sanitize_filename(lecture_title))}
    if kind == "external_link":
        return {"type": kind, "file": f"{asset['filename']}.url", "url": asset["download_url"]}
    return {"type": kind, "file": asset["filename"], "url": asset["download_url"]}


def lecture_plan(udemy, lecture):
    lecture_title = lecture["lecture_title"]
    plan = {"index": lecture["index"], "id": lecture["id"], "class": lecture["_class"], "title": lecture_title}
    if lecture["_class"] in QUIZ_LIKE_CLASSES:
        plan["kind"] = "quiz"
        plan["file"] = sanitize_filename(lecture_title + ".html")
        return plan

    udemy.session.variant_urls = []
    parsed = udemy._parse_lecture(copy.deepcopy(lecture))
    attach_variant_urls(parsed.get("sources") or [], udemy.session.variant_urls)
    extension = parsed.get("extension") or "mp4"
    if parsed.get("is_encrypted"):
        plan["kind"] = "drm"
        plan["file"] = deEmojify(sanitize_filename(lecture_title + "." + extension))
    elif extension == "html":
        plan["kind"] = "html"
        plan["file"] = "{}.html".format(sanitize_filename(lecture_title))
        plan["has_body"] = bool(parsed.get("html_content"))
    else:
        plan["kind"] = "video"
        plan["file"] = deEmojify(sanitize_filename(lecture_title + "." + extension))
        plan["choices"] = {str(quality): choose(parsed.get("sources") or [], quality) for quality in QUALITIES}

    if parsed.get("subtitles") is not None and parsed.get("extension") is None:
        plan["captions"] = [caption_plan(lecture_title, caption) for caption in parsed["subtitles"]]
    plan["assets"] = [asset_plan(lecture_title, asset) for asset in parsed.get("assets") or []]
    return plan


def course_plan(content, hls_dir, continuous_numbers, skip_hls):
    udemy = build_udemy(content, hls_dir, continuous_numbers, skip_hls)
    return {
        "continuous_numbers": continuous_numbers,
        "skip_hls": skip_hls,
        "course_dir": sanitize_filename(udemy.parsed_data["course_title"]),
        "chapters": [
            {
                "index": chapter["chapter_index"],
                "id": chapter["chapter_id"],
                "dir": chapter["chapter_title"],
                "lectures": [lecture_plan(udemy, lecture) for lecture in chapter["lectures"]],
            }
            for chapter in udemy.parsed_data["chapters"]
        ],
    }


def hls_variants(content, hls_dir):
    udemy = build_udemy(content, hls_dir, continuous_numbers=False, skip_hls=False)
    variants = {}
    for master_url in udemy.session.masters:
        udemy.session.variant_urls = []
        sources = udemy._extract_m3u8(master_url)
        attach_variant_urls(sources, udemy.session.variant_urls)
        variants[master_url] = [{"width": source["width"], "height": source["height"], "uri": source["variant_url"]} for source in sources]
    return variants


def main():
    content_path, hls_dir, golden_path, hls_golden_path = (os.path.abspath(arg) for arg in sys.argv[1:5])
    with open(content_path, encoding="utf-8") as handle:
        content = json.load(handle)
    os.chdir(tempfile.mkdtemp(prefix="plan-golden-"))
    plans = [
        course_plan(content, hls_dir, continuous_numbers=False, skip_hls=False),
        course_plan(content, hls_dir, continuous_numbers=True, skip_hls=True),
    ]
    with open(golden_path, "w", encoding="utf-8") as handle:
        json.dump(plans, handle, ensure_ascii=False, indent=1)
        handle.write("\n")
    with open(hls_golden_path, "w", encoding="utf-8") as handle:
        json.dump(hls_variants(content, hls_dir), handle, ensure_ascii=False, indent=1)
        handle.write("\n")
    print(f"wrote {sum(len(chapter['lectures']) for chapter in plans[0]['chapters'])} lecture plans per variant")


if __name__ == "__main__":
    main()
