#!/bin/sh
for argument in "$@"; do printf "%s\n" "$argument"; done
[ "$1" = "https://cdn.example/fails" ] && exit 1
exit 0
