# Docker 部署指南

## 前置要求

- [Docker](https://docs.docker.com/get-docker/) ≥ 20.10
- [Docker Compose](https://docs.docker.com/compose/install/) ≥ 2.0（推荐）

## 快速启动（docker compose）

```bash
git clone https://github.com/dezhishen/now-and-again.git
cd now-and-again

# 启动服务（后台运行）
docker compose up -d

# 查看日志
docker compose logs -f

# 查看状态
docker compose ps
```

服务启动后访问 `http://localhost:8080`。

> 首次启动会自动完成：数据库建表、种子数据、JWT 密钥生成、默认管理员创建。

## 手动构建

```bash
# 构建后端镜像（含前端制品）
docker build -t now-and-again .

# 构建 CLI 镜像
docker build -t now-and-again-cli -f cli/Dockerfile .

# 运行
docker run -d \
  --name now-and-again \
  -p 8080:8080 \
  -v na-data:/data \
  -e GIN_MODE=release \
  now-and-again
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DATA_DIR` | `/data` | 数据根目录（数据库、上传文件、日志、JWT 密钥） |
| `GIN_MODE` | `debug` | 运行模式，生产环境设 `release` |
| `PORT` | `8080` | HTTP 监听端口 |
| `JWT_SECRET` | (自动生成) | JWT 签名密钥，不设置则自动生成并持久化 |
| `ADMIN_DEFAULT_PASSWORD` | (随机生成) | 首次启动时的管理员密码 |

## 数据持久化

所有数据存放在 `DATA_DIR`（默认 `/data`）下：

```
/data/
├── now-and-again.db    # SQLite 数据库
├── .jwt_secret         # JWT 签名密钥（自动生成）
├── uploads/            # 上传文件
└── logs/               # 运行日志
```

使用 Docker 命名卷持久化：

```yaml
volumes:
  - na-data:/data
```

或挂载到宿主机目录：

```yaml
volumes:
  - ./na-data:/data
```

## 默认管理员

首次启动时自动创建管理员账户：

- 用户名：`admin`
- 密码：由 `ADMIN_DEFAULT_PASSWORD` 环境变量设置，未设置则随机生成并打印到容器日志

```bash
# 查看生成的密码
docker compose logs | grep -i "admin password"
```



## 健康检查

镜像内置健康检查：`curl http://localhost:8080/api/system/status`

```bash
# 查看健康状态
docker inspect --format '{{.State.Health.Status}}' now-and-again
```

## 升级

```bash
# 拉取最新代码
git pull

# 重新构建并重启
docker compose down
docker compose up -d --build
```

> 数据库迁移在启动时自动执行（GORM AutoMigrate），无需手动操作。

## 故障排查

### 端口冲突

```bash
# 修改宿主机端口映射
docker run -p 3000:8080 ...
```

### 查看容器日志

```bash
docker compose logs -f
# 或
docker logs -f now-and-again
```

### 进入容器

```bash
docker compose exec server sh
```

### 检查数据目录

```bash
docker compose exec server ls -la /data
docker compose exec server cat /data/.jwt_secret
```
