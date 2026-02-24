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
echo "231964a9a5bf74f9fb615c1bdcae97a9"
md5sum dist/personalities.json

echo "=== Fixing The King ==="
bash ./scripts/buildKing.sh
echo "king hash check, valid and actual:"
echo "489bdc755fe6539b0f9f71b36fea97fb"
md5sum dist/TheKing350noOpk.exe

echo "=== Building Runbook ==="
bash ./scripts/buildRunbook.sh
echo "runbook hash check, valid and actual:"
echo "32f6c13255833cdd7995a50b689b3e1b"
md5sum dist/runbook
