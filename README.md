# Cloud Drive

Cloud Drive 是一个全栈网盘示例项目，包含：

- 前端：Vue 3 + TypeScript（`apps/cloud-drive-frontend`）
- 后端：Go + Gin（`apps/cloud-drive-backend`）
- 基础设施：MySQL（`docker compose`）

## 前置要求

- Node.js 20+
- `pnpm` 9+
- Docker Desktop（含 `docker compose`）

## 一键命令速查

以下命令均在仓库根目录执行：`/Users/shi/study/frontend/projects/cloud-drive`。

### 1) 一键启动

- 启动基础设施（MySQL）：

```bash
docker compose --profile infra up -d
```

- 启动后端链路（MySQL + Backend）：

```bash
docker compose --profile backend up -d --build
```

- 启动全栈（MySQL + Backend + Frontend）：

```bash
docker compose --profile full up -d --build
```

### 2) 一键停止

```bash
docker compose down
```

### 3) 一键重建（不清数据）

重建容器和镜像，但保留 `./data/mysql` 持久化数据。

```bash
docker compose --profile full up -d --build --force-recreate
```

### 4) 一键恢复（依赖持久化目录）

当你已经有历史数据目录时，直接拉起服务即可恢复：

```bash
docker compose --profile infra up -d
```

可选验证：

```bash
docker compose exec -T mysql mysql -uroot -p123456123456 -D cloud-drive -e "SHOW TABLES;"
```

### 5) 一键导入外部 SQL

推荐用封装脚本：

```bash
./scripts/mysql-import.sh ./ops/mysql/import/your_dump.sql cloud-drive
```

也可直接执行：

```bash
docker compose exec -T mysql \
  mysql -uroot -p123456123456 cloud-drive < ./ops/mysql/import/your_dump.sql
```

### 6) 一键清理

- 仅停服务（保留数据）：

```bash
docker compose down
```

- 停服务并删除容器匿名卷（仍保留 `./data/*`）：

```bash
docker compose down -v
```

- 彻底清空本项目数据库数据（危险操作）：

```bash
docker compose down
find data/mysql -mindepth 1 -maxdepth 1 -exec rm -rf {} +
```

## 开发运行（本地非容器）

```bash
pnpm install
cd apps/cloud-drive-backend && go run cmd/server/main.go
cd apps/cloud-drive-frontend && pnpm dev
```

## 排障说明

### 端口冲突

- 现象：`bind: address already in use`（如 `3306` 被占用）。
- 解决：临时改映射端口启动。

```bash
MYSQL_PORT=13306 docker compose --profile infra up -d
```

### MySQL 未就绪导致导入失败

- 现象：`MySQL container is not running` 或连接失败。
- 解决：先启动并等待健康检查通过，再执行导入。

```bash
docker compose --profile infra up -d
docker compose ps
```

### 初始化 SQL 没执行

- 原因：MySQL 只有在 `./data/mysql` 为空时才会自动执行 `ops/mysql/init/*.sql`。
- 解决：清空 `data/mysql` 后冷启动，或手动执行初始化 SQL。

## Task5 验证记录（实测）

以下为本次在本机执行并通过的验证（由于 `3306` 冲突，使用了临时端口环境变量）。

### A. 冷启动验证

命令：

```bash
docker compose down
mkdir -p data/mysql
find data/mysql -mindepth 1 -maxdepth 1 -exec rm -rf {} +
MYSQL_PORT=13306 docker compose --profile infra up -d
MYSQL_PORT=13306 docker compose exec -T mysql \
  mysql -uroot -p123456123456 -D cloud-drive \
  -e "SHOW TABLES; SELECT * FROM bootstrap_records;"
```

结果摘要：

- `bootstrap_records` 表存在
- 数据包含 `mysql-initdb-loaded` 记录，说明初始化 SQL 生效

### B. 重建恢复验证

命令：

```bash
MYSQL_PORT=13306 docker compose exec -T mysql \
  mysql -uroot -p123456123456 -D cloud-drive \
  -e "INSERT INTO bootstrap_records(tag) VALUES('rebuild-keep') ON DUPLICATE KEY UPDATE tag=VALUES(tag);"
MYSQL_PORT=13306 docker compose down
MYSQL_PORT=13306 docker compose --profile infra up -d --force-recreate
MYSQL_PORT=13306 docker compose exec -T mysql \
  mysql -uroot -p123456123456 -D cloud-drive \
  -e "SELECT * FROM bootstrap_records ORDER BY id;"
```

结果摘要：

- 重建后 `rebuild-keep` 记录仍在
- 证明 MySQL 数据卷恢复正常

### C. 外部导入验证

命令：

```bash
MYSQL_PORT=13306 ./scripts/mysql-import.sh \
  ./ops/mysql/import/manual_verify.sql cloud-drive
MYSQL_PORT=13306 docker compose exec -T mysql \
  mysql -uroot -p123456123456 -D cloud-drive \
  -e "SELECT * FROM manual_import_records ORDER BY id DESC LIMIT 3;"
```

结果摘要：

- 导入脚本输出 `Import completed.`
- 查询到 `manual-import-ok` 记录，外部导入链路可用
