#!/bin/sh
# 修复数据目录权限（处理 Docker 卷挂载的权限问题）
if [ -d "/app/data" ]; then
    chown -R appuser:appgroup /app/data
fi

# 切换到 appuser 运行服务
exec su-exec appuser:appgroup /app/server "$@"
