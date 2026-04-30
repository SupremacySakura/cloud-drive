#!/bin/bash

set -e

RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

STAGED_ENV_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.env($|\.)' | grep -v '\.env\.example$' || true)

if [ -z "$STAGED_ENV_FILES" ]; then
  exit 0
fi

echo -e "${YELLOW}⚠️  警告: 检测到以下 .env 文件将被提交:${NC}"
for file in $STAGED_ENV_FILES; do
  echo "   - $file"
done

echo -e "${RED}❌ 提交被拒绝: 请勿提交 .env 文件 (包含敏感信息)${NC}"
echo "   如需强制提交，使用: git commit --no-verify"

exit 1
