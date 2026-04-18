# 环境变量配置指南

本文档说明如何配置云盘后端服务的环境变量。

## 快速开始

1. 复制示例配置文件：
   ```bash
   cp .env.example .env
   ```

2. 编辑 `.env` 文件，修改为你自己的配置

3. 启动服务：
   ```bash
   go run cmd/server/main.go
   ```

## 配置项说明

### 服务器配置

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PORT` | `9000` | 服务器监听端口 |

### 数据库配置 (MySQL)

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `DB_USER` | `root` | 数据库用户名 |
| `DB_PASSWORD` | `123456123456` | 数据库密码 |
| `DB_HOST` | `127.0.0.1` | 数据库主机地址 |
| `DB_PORT` | `3306` | 数据库端口 |
| `DB_NAME` | `cloud-drive` | 数据库名称 |

### JWT 认证配置

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `JWT_SECRET` | `your_secret_key` | JWT 签名密钥 |

**⚠️ 安全警告**：生产环境必须修改 `JWT_SECRET`！

生成强密钥的命令：
```bash
openssl rand -base64 32
```

### 文件存储配置

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `CHUNK_STORAGE_PATH` | `./data` | 分片上传临时存储路径 |
| `FILE_STORAGE_PATH` | `./data` | 合并后文件存储路径 |

## 优先级

配置加载优先级（从高到低）：

1. 操作系统环境变量
2. `.env` 文件中的变量
3. 代码中的默认值

## 生产环境部署

### Docker 部署示例

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 9000
CMD ["./server"]
```

运行容器时传入环境变量：

```bash
docker run -d \
  -e JWT_SECRET="your-production-secret" \
  -e DB_PASSWORD="your-db-password" \
  -e DB_HOST="mysql-host" \
  -p 9000:9000 \
  cloud-drive-backend
```

### Systemd 服务示例

创建 `/etc/systemd/system/cloud-drive.service`：

```ini
[Unit]
Description=Cloud Drive Backend
After=network.target

[Service]
Type=simple
User=cloud-drive
WorkingDirectory=/opt/cloud-drive
Environment="JWT_SECRET=your-production-secret"
Environment="DB_PASSWORD=your-db-password"
Environment="DB_HOST=localhost"
Environment="FILE_STORAGE_PATH=/var/cloud-drive/data"
ExecStart=/opt/cloud-drive/server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

启用并启动服务：

```bash
sudo systemctl enable cloud-drive
sudo systemctl start cloud-drive
```

## 安全最佳实践

1. **永远不要提交 `.env` 文件到 Git**
   - `.env` 已添加到 `.gitignore`
   - 只提交 `.env.example` 作为模板

2. **生产环境使用强密码和密钥**
   - 数据库密码至少 16 位随机字符
   - JWT 密钥使用 `openssl rand -base64 32` 生成

3. **限制文件存储目录权限**
   ```bash
   chmod 750 /var/cloud-drive/data
   chown cloud-drive:cloud-drive /var/cloud-drive/data
   ```

4. **使用 HTTPS**
   - 生产环境必须配置 TLS
   - 建议使用反向代理（Nginx/Caddy）

5. **定期轮换密钥**
   - JWT 密钥应定期更换
   - 更换后所有用户需要重新登录

## 故障排查

### 数据库连接失败

检查环境变量：
```bash
echo $DB_HOST $DB_PORT $DB_USER
```

确认 MySQL 服务运行：
```bash
mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p
```

### JWT 认证失败

检查 JWT_SECRET 是否设置：
```bash
echo $JWT_SECRET
```

如果密钥已更改，所有用户需要重新登录获取新 token。

### 文件上传失败

检查存储路径权限：
```bash
ls -la $FILE_STORAGE_PATH
```

确保服务用户有读写权限。
