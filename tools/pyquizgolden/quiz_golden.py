"""Record how the Python downloader serializes JSON and writes quiz pages, for the Go port's golden tests.

Usage: quiz_golden.py <pyjson_golden.json> <quiz_fixture_dir> [--existing]
Run with the udemy-downloader venv and PYTHONPATH set to that repo. The synthetic fixtures below are
written into the fixture dir first; with --existing, only the <name>.json files already there (for
example quizzes captured from a real course) are rendered.
"""

import json
import logging
import sys
import tempfile
from pathlib import Path

from pathvalidate import sanitize_filename

from udemy_downloader.Udemy import Udemy

JSON_CASES = [
    '{"b": 1, "a": [true, false, null], "c": {}}',
    '[]',
    '{"float": [0.1, 1.5, 100.0, 1e16, 1e15, 0.0001, 0.00001, 1.5e-07, -0.0, 2.5e+20, 123456789012345678.0]}',
    '{"int": [0, -0, 7, 12345678901234567890123, -42]}',
    '{"overflow": [1e400, -1e400]}',
    '"Tiếng Việt — emoji 😀 and \\u0000 \\u001f \\u007f tab\\t quote\\" slash/ backslash\\\\"',
    '{"dup": 1, "other": 2, "dup": 3}',
    '{"script": "</script><script>alert(1)</script>"}',
]

LECTURE_DATA = {"_class": "quiz", "id": 901, "title": "Quiz 1: Cloud & \"AI\" <b>'Tiếng Việt'</b> 😀", "object_index": 7, "type": "simple-quiz"}

QUIZ_FIXTURES = {
    "normal_quiz": {
        "lecture": {"id": 901, "lecture_title": "007 Quiz 1 Cloud & AI", "lecture_index": 7, "_class": "quiz", "data": LECTURE_DATA},
        "assessments": [
            {"_class": "assessment", "id": 1, "assessment_type": "multiple-choice",
             "prompt": {"question": "<p>Which service runs </script> code? ✓</p>", "answers": ["<p>Lambda</p>", "<p>S3 – storage</p>"],
                        "feedbacks": ["Đúng rồi", ""], "relatedLectureIds": ""},
             "correct_response": ["a"], "section": "Compute", "question_plain": "Which service runs code?",
             "related_lectures": [{"id": 5, "title": "Intro"}], "score": 0.1, "weight": 1e-07},
            {"_class": "assessment", "id": 2, "assessment_type": "multi-select",
             "prompt": {"question": "Pick 😀 two", "answers": ["A", "B", "C"]}, "correct_response": ["a", "c"],
             "section": None, "question_plain": "Pick two", "related_lectures": []},
        ],
    },
    "quiz_with_description": {
        "lecture": {"id": 902, "lecture_title": "008 Practice test", "lecture_index": 8, "_class": "quiz",
                    "data": {**LECTURE_DATA, "id": 902, "title": "Practice test", "description": "<p>Mô tả</p>", "pass_percent": 72}},
        "assessments": [{"_class": "assessment", "id": 3, "assessment_type": "multiple-choice", "prompt": {"question": "Q"}}],
    },
    "coding_assignment": {
        "lecture": {"id": 903, "lecture_title": "009 Coding exercise", "lecture_index": 9, "_class": "quiz", "data": {**LECTURE_DATA, "id": 903}},
        "assessments": [{"_class": "assessment", "id": 4, "assessment_type": "coding-problem",
                         "prompt": {"instructions": "<p>Write </script> a func</p>",
                                    "test_files": [{"file_name": "test.py", "content": "assert f() == 1\n"}],
                                    "solution_files": [{"file_name": "f.py", "content": "def f():\n    return 1\n"}]}}],
    },
    "coding_assignment_missing_parts": {
        "lecture": {"id": 904, "lecture_title": "010 Coding exercise 2", "lecture_index": 10, "_class": "quiz", "data": {**LECTURE_DATA, "id": 904}},
        "assessments": [{"_class": "assessment", "id": 5, "assessment_type": "coding-problem",
                         "prompt": {"instructions": "(None)", "test_files": None}}],
    },
    "coding_assignment_without_prompt": {
        "lecture": {"id": 905, "lecture_title": "011 Coding exercise 3", "lecture_index": 11, "_class": "quiz", "data": {**LECTURE_DATA, "id": 905}},
        "assessments": [{"_class": "assessment", "id": 6, "assessment_type": "coding-problem", "prompt": {}}],
    },
    "coding_problem_among_others": {
        "lecture": {"id": 906, "lecture_title": "012 Mixed", "lecture_index": 12, "_class": "quiz", "data": {**LECTURE_DATA, "id": 906}},
        "assessments": [{"_class": "assessment", "id": 7, "assessment_type": "coding-problem", "prompt": {}},
                        {"_class": "assessment", "id": 8, "assessment_type": "multiple-choice", "prompt": {}}],
    },
    "empty_quiz": {
        "lecture": {"id": 907, "lecture_title": "013 Empty", "lecture_index": 13, "_class": "quiz", "data": {**LECTURE_DATA, "id": 907}},
        "assessments": [],
    },
    "assessment_without_type": {
        "lecture": {"id": 908, "lecture_title": "014 Untyped", "lecture_index": 14, "_class": "quiz", "data": {**LECTURE_DATA, "id": 908}},
        "assessments": [{"_class": "assessment", "id": 9, "prompt": {}}],
    },
}

RECORDED_SUFFIXES = (".html", ".error")


def write_json_golden(path: Path) -> None:
    cases = [{"input": case, "output": json.dumps(json.loads(case))} for case in JSON_CASES]
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(cases, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def quiz_downloader(assessments) -> Udemy:
    udemy = Udemy.__new__(Udemy)
    udemy.logger = logging.getLogger("quiz-golden")
    udemy._get_all_pages = lambda url, params=None, description="data": {"results": assessments}
    udemy.portal_name = "example-portal"
    return udemy


def render(fixture_dir: Path, name: str) -> None:
    for suffix in RECORDED_SUFFIXES:
        (fixture_dir / f"{name}{suffix}").unlink(missing_ok=True)
    fixture = json.loads((fixture_dir / f"{name}.json").read_text(encoding="utf-8"))
    lecture = fixture["lecture"]
    with tempfile.TemporaryDirectory() as chapter_dir:
        try:
            quiz_downloader(fixture["assessments"]).process_quiz(lecture, Path(chapter_dir))
        except Exception as error:  # the downloader records any failure as a failed quiz
            (fixture_dir / f"{name}.error").write_text(type(error).__name__ + "\n", encoding="utf-8")
            return
        page = Path(chapter_dir, sanitize_filename(lecture["lecture_title"] + ".html")).read_bytes()
    (fixture_dir / f"{name}.html").write_bytes(page)


def main(json_golden: Path, fixture_dir: Path, write_synthetic: bool) -> None:
    write_json_golden(json_golden)
    fixture_dir.mkdir(parents=True, exist_ok=True)
    if write_synthetic:
        for name, fixture in QUIZ_FIXTURES.items():
            (fixture_dir / f"{name}.json").write_text(json.dumps(fixture, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    for fixture in sorted(fixture_dir.glob("*.json")):
        render(fixture_dir, fixture.stem)


if __name__ == "__main__":
    main(Path(sys.argv[1]), Path(sys.argv[2]), write_synthetic="--existing" not in sys.argv[3:])
