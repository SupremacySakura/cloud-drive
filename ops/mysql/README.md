# MySQL persistence and import guide

## 1) Persistent volumes

`docker-compose.yml` mounts:

- `./data/mysql` -> `/var/lib/mysql` (MySQL data persistence)
- `./ops/mysql/init` -> `/docker-entrypoint-initdb.d` (MySQL first-start init scripts)
- `./ops/mysql/import` -> `/opt/mysql-import` (manual import staging dir)

## 2) MySQL initialization import

On first MySQL startup (empty `./data/mysql`), MySQL auto-runs all SQL files in
`./ops/mysql/init`.

Current init script:

- `./ops/mysql/init/001_bootstrap.sql`

The MySQL entrypoint creates `MYSQL_DATABASE`. The script creates
`bootstrap_records` in that selected database and inserts an idempotent marker.

## 启动与升级约束

- 两种启动方式的完整步骤见根目录 [README](../../README.md#两种启动方式)。
- 全栈 Docker 使用根 `.env`、`data/mysql` 与现有文件卷；Run dev 使用 `.env.development` 和本机 MySQL 服务（3306），文件存储为 `data/dev/files`，不依赖 Docker。Docker MySQL 默认宿主机端口为 3307。
- MySQL 官方 entrypoint 创建并选择 `MYSQL_DATABASE` 后才执行初始化 SQL；`001_bootstrap.sql` 只在当前库创建标记表，不硬编码库名。
- 应用表由后端 `AutoMigrate` 创建；三个文件查询索引在表创建后由后端幂等补建，已有索引不会重复创建。
- 旧 `cloud-drive` 库无需更名：保留 `.env` 中 `MYSQL_DATABASE=cloud-drive` 和 `DB_NAME=cloud-drive`，并保留原密码、数据目录和卷名。不要删除数据来解决库名不一致。

### 主动更名数据库（可选）

更名不是启动前置条件，生产环境应先安排停写窗口。停止前后端但保留 MySQL，备份旧库和文件卷，再创建并导入新库。以下命令使用容器中的 root 密码，备份只包含旧库，不写入线上新库：

```bash
docker compose stop frontend backend
docker compose exec -T mysql sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysqldump -uroot --single-transaction --routines --triggers --events --no-tablespaces --set-gtid-purged=OFF cloud-drive' > /tmp/zhixu-legacy-backup.sql
```

确认备份成功且有文件卷备份后，再由管理员创建 `zhixu` 并将备份导入目标库（导入时指定目标库，不使用带旧库 `USE` 的 dump）。核对表行数、登录与文件关联后，才把 `.env` 两个库名切换为 `zhixu` 并启动服务。
保留原库与备份；回滚时停止写入、恢复对应数据库和文件备份再切换原配置，避免切换后新写入丢失。不要直接在旧数据目录上清空重建。

## 3) Manual external import

### Option A: import from any local SQL file (recommended)

```bash
./scripts/mysql-import.sh ./ops/mysql/import/your_dump.sql zhixu
```

### Option B: direct docker compose command

```bash
docker compose exec -T mysql \
  mysql -uroot -p123456123456 zhixu < ./ops/mysql/import/your_dump.sql
```

## 4) Quick verify commands

```bash
docker compose up -d mysql
docker compose exec -T mysql mysql -uroot -p123456123456 -D zhixu -e "SHOW TABLES;"
docker compose exec -T mysql mysql -uroot -p123456123456 -D zhixu -e "SELECT * FROM bootstrap_records;"
```
