#!/bin/sh
events=""
previous=""
for argument in "$@"; do
  [ "$previous" = "--events-json" ] && events="$argument"
  previous="$argument"
done
echo "argv: $*"
echo "token present: ${UDEMY_ACCESS_TOKEN:+yes}"
echo "working dir: $(pwd)"
exec 4>"$events"
case "$FAKE_PYTHON_MODE" in
  success)
    printf '%s\n' '{"event": "course_started", "course_title": "Backend", "total_lectures": 2}' >&4
    printf '%s\n' '{"event": "lecture_started", "chapter_index": 1, "lecture_index": 1, "title": "001 A"}' >&4
    printf '%s\n' '{"event": "lecture_done", "chapter_index": 1, "lecture_index": 1, "title": "001 A", "status": "downloaded"}' >&4
    printf '%s\n' '{"event": "lecture_failed", "chapter_index": 1, "lecture_index": 2, "title": "002 B", "failures": ["002 B (caption: en)"]}' >&4
    printf '%s\n' '{"event": "summary", "downloaded": 1, "already_present": 0, "failed": ["002 B (caption: en)"]}' >&4
    ;;
  failure)
    echo "Traceback (most recent call last): login rejected" >&2
    exit 1
    ;;
  wait-for-interrupt)
    sh -c 'trap "" INT; while :; do sleep 0.05; done' &
    echo $! > "$FAKE_CHILD_PID_FILE"
    trap 'echo "interrupted"; exit 130' INT
    printf '%s\n' '{"event": "course_started", "course_title": "Backend", "total_lectures": 9}' >&4
    while :; do sleep 0.05; done
    ;;
esac
