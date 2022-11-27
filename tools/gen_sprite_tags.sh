#!/bin/bash
if [[ $1 == "-h" ]]; then
	echo "this script extracts frame tags from an aseprite-exported sprite JSON and generates Go constants to refer to them by name" >&2
fi

if [[ $# -ne 3 ]]; then
	echo "this is script needs exactly 3 arguments: input file and output file and const prefix" >&2
	exit 1
fi

echo -e "package main\n\ntype $3AnimationTags uint8\n\nconst (" > "$2"
jq -r '.meta.frameTags[].name' "$1" \
	| sed 's/ //' \
	| sed "1s/$/ $3AnimationTags = iota/" \
	| sed "s/^/$3/" \
	>> "$2"
echo ")" >> "$2"
gofmt -s -w "$2"
