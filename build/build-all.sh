#!/bin/bash
set -e

echo "=== Building BIN books ==="
bash ./scripts/books2bin.sh
echo "Book bin folder hash check, valid and actual:"
echo "55fe7c8c118574aa7fde1f713dce046d"
md5=$(cd /build/dist/books && md5sum * | md5sum | head -c 32)
echo "$md5"

echo "=== Building personalities.json ==="
node ./scripts/personalitiesBuilder.js
echo "personalities.json hash check, valid and actual:"
echo "17e79cab47a4938f1428e2185002a20c"
md5sum dist/personalities.json

echo "=== Fixing The King ==="
bash ./scripts/buildKing.sh
echo "king hash check, valid and actual:"
echo "d858f0015870c458431ce79175d127de"
md5sum dist/TheKing350noOpk.exe
