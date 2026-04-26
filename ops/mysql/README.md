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

It creates database `cloud-drive`, creates table `bootstrap_records`, and inserts
an idempotent bootstrap marker record.

## 3) Manual external import

### Option A: import from any local SQL file (recommended)

```bash
./scripts/mysql-import.sh ./ops/mysql/import/your_dump.sql cloud-drive
```

### Option B: direct docker compose command

```bash
docker compose exec -T mysql \
  mysql -uroot -p123456123456 cloud-drive < ./ops/mysql/import/your_dump.sql
```

## 4) Quick verify commands

```bash
docker compose up -d mysql
docker compose exec -T mysql mysql -uroot -p123456123456 -D cloud-drive -e "SHOW TABLES;"
docker compose exec -T mysql mysql -uroot -p123456123456 -D cloud-drive -e "SELECT * FROM bootstrap_records;"
```
