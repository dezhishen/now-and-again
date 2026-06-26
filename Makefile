.PHONY: help dev dev-backend dev-frontend dev-cli \
        build build-backend build-cli \
        test test-backend test-shared \
        lint lint-backend lint-frontend \
        db-reset db-seed \
        install install-cli \
        clean check-contracts fix-dupes \
        docker-build docker-up docker-down docker-logs

# ─── Default ──────────────────────────────────────────────────────
.DEFAULT_GOAL := help

help: ## 显示所有可用目标
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}'

# ─── Development ──────────────────────────────────────────────────

dev: ## 并行启动 backend + frontend（使用 Ctrl+C 同时停止）
	@echo "→ Backend  http://localhost:8080"
	@echo "→ Frontend http://localhost:5173"
	@cd backend && ADMIN_DEFAULT_PASSWORD=12345678 DATA_DIR=../data go run ./cmd/server & \
		cd frontend && pnpm run dev & \
		wait

dev-backend: ## 启动后端开发服务器 (:8080)
	@echo "→ Backend  http://localhost:8080"
	@cd backend && ADMIN_DEFAULT_PASSWORD=12345678 DATA_DIR=../data go run ./cmd/server

dev-frontend: ## 启动前端开发服务器 (:5173)
	@echo "→ Frontend http://localhost:5173"
	@cd frontend && pnpm run dev

dev-cli: ## 编译并运行 CLI（参数通过 ARGS 传递，如 make dev-cli ARGS="task list --family-id xxx"）
	@cd cli && go run . $(ARGS)

# ─── Build ────────────────────────────────────────────────────────

build: build-backend build-cli build-frontend ## 构建全部

build-backend: ## 编译后端二进制 → backend/na-server
	@echo "Building backend..."
	@cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o na-server ./cmd/server
	@echo "  → backend/na-server"

build-cli: ## 编译 CLI 二进制 → cli/na
	@echo "Building CLI..."
	@cd cli && CGO_ENABLED=0 go build -ldflags="-s -w" -o na .
	@echo "  → cli/na"

build-frontend: ## 构建前端 → frontend/dist/
	@echo "Building frontend..."
	@cd frontend && pnpm install --frozen-lockfile && pnpm run build
	@echo "  → frontend/dist/"

install-cli: build-cli ## 安装 CLI 到 $GOPATH/bin 或 /usr/local/bin
	@if [ -w /usr/local/bin ]; then \
		cp cli/na /usr/local/bin/na && echo "→ /usr/local/bin/na"; \
	else \
		cp cli/na $(shell go env GOPATH)/bin/na && echo "→ $(shell go env GOPATH)/bin/na"; \
	fi

# ─── Test ─────────────────────────────────────────────────────────

test: test-backend test-shared ## 运行所有 Go 测试

test-backend: ## 运行 backend 测试
	@cd backend && go test ./... -count=1 -short

test-shared: ## 运行 shared 模块测试
	@cd shared && go test ./... -count=1 -short

test-cli: ## 运行 CLI 测试
	@cd cli && go test ./... -count=1 -short

# ─── Lint / Vet ───────────────────────────────────────────────────

lint: lint-backend lint-frontend ## 运行所有代码检查

lint-backend: ## Go vet backend + shared + CLI
	@echo "→ vet shared/"
	@cd shared && go vet ./...
	@echo "→ vet backend/"
	@cd backend && go vet ./...
	@echo "→ vet cli/"
	@cd cli && go vet ./...

lint-frontend: ## TypeScript 类型检查
	@echo "→ vue-tsc"
	@cd frontend && pnpm exec vue-tsc --noEmit

fix-dupes: ## 自动修复 Go 文件中重复的 package 声明（VSCode 自动补全副作用）
	@for f in $$(find . -name '*.go' -not -path './.git/*'); do \
		count=$$(grep -c '^package ' "$$f" 2>/dev/null || echo 0); \
		if [ "$$count" -gt 1 ]; then \
			sed -i '1{/^package /d}' "$$f"; \
			echo "  fixed: $$f"; \
		fi; \
	done

check-contracts: fix-dupes ## 验证 backend 和 CLI 都实现了 contracts 接口（编译检查）
	@echo "→ Checking backend implements contracts..."
	@cd backend && go build ./... || (echo "❌ backend 编译失败 — 检查是否实现了所有 contracts 接口" && exit 1)
	@echo "  ✅ backend"
	@echo "→ Checking CLI implements contracts..."
	@cd cli && go build ./... || (echo "❌ CLI 编译失败 — 检查是否实现了所有 contracts 接口" && exit 1)
	@echo "  ✅ CLI"

# ─── Database ─────────────────────────────────────────────────────

db-reset: ## 删除 SQLite 数据库文件
	@rm -f data/*.db data/*.db-journal data/*.db-wal data/*.db-shm
	@echo "→ database removed"

db-seed: ## 仅运行种子数据（需先启动后端）
	@echo "Seed data is applied on server startup (AutoMigrate + Seed)."
	@echo "Simply restart the backend: make dev-backend"

# ─── Clean ────────────────────────────────────────────────────────

clean: ## 清理所有构建产物
	@echo "Cleaning..."
	@rm -f backend/na-server
	@rm -f cli/na
	@rm -rf frontend/dist/
	@rm -rf frontend/.vite/
	@rm -f backend/*.db backend/*.db-journal backend/*.db-wal backend/*.db-shm
	@echo "→ done"

# ─── Dependencies ─────────────────────────────────────────────────

deps: ## 安装所有依赖
	@echo "→ Go modules (shared)..."
	@cd shared && go mod tidy
	@echo "→ Go modules (backend)..."
	@cd backend && go mod tidy
	@echo "→ Go modules (CLI)..."
	@cd cli && go mod tidy
	@echo "→ pnpm (frontend)..."
	@cd frontend && pnpm install
	@echo "→ all dependencies ready"

# ─── CI ───────────────────────────────────────────────────────────

ci: deps check-contracts lint test build ## CI 完整流水线
	@echo "→ CI passed ✅"

# ─── Docker ───────────────────────────────────────────────────────

docker-build: ## 构建 Docker 镜像
	docker build -t now-and-again-backend -f backend/Dockerfile .
	docker build -t now-and-again-frontend -f frontend/Dockerfile .

docker-up: ## 启动 docker-compose
	docker compose up -d

docker-down: ## 停止 docker-compose
	docker compose down

docker-logs: ## 查看 docker-compose 日志
	docker compose logs -f
