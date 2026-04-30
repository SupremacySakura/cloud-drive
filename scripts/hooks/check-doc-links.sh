#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [ ! -d "$REPO_ROOT/docs" ]; then
  echo -e "${YELLOW}⚠️  docs/ 目录不存在，跳过链接检查${NC}"
  exit 0
fi

BROKEN_LINKS=0

for md_file in "$REPO_ROOT"/docs/**/*.md; do
  if [ ! -f "$md_file" ]; then
    continue
  fi

  while IFS= read -r link; do
    [ -z "$link" ] && continue

    if [[ "$link" == http* ]] || [[ "$link" == https* ]]; then
      continue
    fi

    if [[ "$link" == \#* ]]; then
      continue
    fi

    if [[ "$link" == /* ]]; then
      target_path="$REPO_ROOT$link"
    else
      dir=$(dirname "$md_file")
      target_path="$dir/$link"
    fi

    target_path=$(cd "$(dirname "$target_path")" 2>/dev/null && pwd)/$(basename "$target_path") 2>/dev/null || target_path="$target_path"

    if [ ! -e "$target_path" ]; then
      rel_file="${md_file#$REPO_ROOT/}"
      echo -e "${RED}❌ 断链: $rel_file -> $link${NC}"
      BROKEN_LINKS=$((BROKEN_LINKS + 1))
    fi
  done < <(grep -oE '\]\([^)]+\)' "$md_file" 2>/dev/null | sed 's/]\(([^)]*)\)/\1/' | sed 's/(//;s/)//')
done

if [ "$BROKEN_LINKS" -gt 0 ]; then
  echo -e "${RED}发现 $BROKEN_LINKS 个断链${NC}"
  exit 1
fi

echo -e "${GREEN}✅ 所有 docs/ 链接检查通过${NC}"
exit 0
