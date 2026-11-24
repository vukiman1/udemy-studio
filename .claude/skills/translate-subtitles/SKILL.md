---
name: translate-subtitles
description: Translate a downloaded course's English subtitles into Vietnamese inside this Claude Code session, at no API cost, and write them into the library only after the app's own checks pass. Use when the user asks to "dịch phụ đề", "dịch khoá", "dịch sub", translate a course's subtitles here, or to continue a translation that stopped part-way.
---

# Translate subtitles in this session

The app can translate subtitles itself with an API key. This skill does the same work here instead, under the user's
Claude subscription, and keeps the app as the gatekeeper: `subtitles export` hands out the English cues and
`subtitles import` writes a translation into the library only when it passes the checks the app runs on its own
translations (every cue present, original timings, no drifted lines).

## Rules

- Never write into the library yourself. Only `subtitles import` writes there.
- The style and format rules live in `internal/translate/system_prompt.txt`, the prompt the app sends to every model.
  Read it before translating and follow it exactly; do not restate or change the rules here.
- Work from the repository root.

## Steps

1. **Pick the course and library.** The library is the `downloads/` directory the app serves (`--library`, or the
   `UDEMY_STUDIO_LIBRARY` environment variable). Ask the user if neither is known. The course is its folder name.
   Use a work directory under `tmp/translate/<course folder>`: `tmp/` is ignored by git and survives between sessions.

2. **Export** the lectures that still lack Vietnamese subtitles:

   ```sh
   go run ./cmd/udemy-studio subtitles export --library "<library>" --course "<course folder>" --work "tmp/translate/<course folder>"
   ```

   It writes `<id>.en.txt` per lecture (one `<index><TAB><text>` line per cue) and `handoff.json`, which lists the
   lectures in course order with their file id. Files are named by a stable lecture id, so exporting again later lists
   only what is left, under the same names, and never pairs an old translation with another lecture.

3. **Read `internal/translate/system_prompt.txt`.** Its voice, keep-in-English terms, speech-recognition fixes and hard
   format rules apply to every line you write. `handoff.json` gives each file's lecture title, which helps restore
   misheard terms.

4. **Translate each `<id>.en.txt`** in the order `handoff.json` lists them, skipping any that already has an `<id>.vi.txt`:
   - Read the whole file, then write `<id>.vi.txt` beside it: exactly one `<index><TAB><Vietnamese text>` line per input
     line, same indices, same order, a real tab character, no empty text, nothing else in the file.
   - For long files, translate in consecutive parts of about 120 lines, but keep the lines just before each part in
     view so sentences that cross the boundary stay consistent. Write one file per lecture.
   - Sentences often span cues: when the English cue stops mid-sentence, stop the Vietnamese cue too and continue in
     the next one. Never move a whole clause into a neighbouring cue.

5. **Import** what is translated:

   ```sh
   go run ./cmd/udemy-studio subtitles import --library "<library>" --work "tmp/translate/<course folder>"
   ```

   It prints the lectures written, rejected (with the reason) and not translated yet, and exits non-zero if any was rejected.

6. **Fix rejections**, then import again:
   - `the translation looks shifted against the original cues: cues X-Y`: lines in that span carry their neighbours'
     meaning. Re-translate that span cue by cue, keeping each cue's own meaning in that cue.
   - `the model's reply does not match the cues it was given`: lines are missing, extra, repeated, empty, or lack the tab.
   - A lecture that already has Vietnamese subtitles is never overwritten; leave it.

7. **Report** how many lectures were written, rejected and left, and remind the user that a running server shows the
   new subtitles after it restarts.

## Large courses

A course can have tens of thousands of cues, which may not fit in one session's usage limits. Translate a batch of
lectures, import them, and tell the user how many are left; the next session repeats steps 2-7 and continues where this
one stopped.
