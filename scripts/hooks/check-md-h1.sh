#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

STAGED_MD_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.md$' || true)

if [ -z "$STAGED_MD_FILES" ]; then
  exit 0
fi

ERRORS=0

for file in $STAGED_MD_FILES; do
  if [ ! -f "$file" ]; then
    continue
  fi

  # Skip files outside docs/ directory
  if [[ ! "$file" =~ ^docs/ ]]; then
    continue
  fi

  H1_COUNT=$(grep -cE '^#[[:space:]]' "$file" 2>/dev/null || echo 0)

  if [ "$H1_COUNT" -eq 0 ]; then
    echo -e "${RED}❌ Error: $file 缺少 H1 标题${NC}"
    ERRORS=$((ERRORS + 1))
  elif [ "$H1_COUNT" -gt 1 ]; then
    echo -e "${RED}❌ Error: $file 有 $H1_COUNT 个 H1 标题，应该只有 1 个${NC}"
    ERRORS=$((ERRORS + 1))
  fi
done

if [ "$ERRORS" -gt 0 ]; then
  echo -e "${RED}发现 $ERRORS 个 Markdown H1 标题问题${NC}"
  exit 1
fi

echo -e "${GREEN}✅ 所有 Markdown 文件 H1 检查通过${NC}"
exit 0
