.PHONY: help dev dev-backend dev-frontend dev-cli \
        build build-backend build-cli build-frontend \
        test test-backend test-cli \
        lint lint-backend lint-frontend \
        db-reset db-seed \
        install install-cli \
	clean check-contracts check-plugin-isolation fix-dupes \
	        docker-build docker-up docker-down docker-logs \
	        docker-e2e-image docker-e2e-image-zh docker-e2e-headed docker-e2e-headed-zh

E2E_IMAGE := now-and-again-e2e
E2E_IMAGE_ZH := now-and-again-e2e:zh
E2E_WORKSPACE := /workspace
E2E_BASE_CMD = docker run --rm -it \
	-v $(PWD):$(E2E_WORKSPACE) \
	-w $(E2E_WORKSPACE) \
	--shm-size=2g \
	--ipc=host \
	-e CI=true

# ─── Default ──────────────────────────────────────────────────────
.DEFAULT_GOAL := help

help: ## 显示所有可用目标
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}'

# ─── Development ──────────────────────────────────────────────────

dev: ## 并行启动 backend + frontend（Ctrl+C 同时停止，自动清理子进程）
	@echo "→ 清理残留进程..."
	@-fuser -k 8080/tcp 2>/dev/null || true
	@-fuser -k 5173/tcp 2>/dev/null || true
	@sleep 0.5
	@mkdir -p .dev
	@LAN_IP=$$(hostname -I 2>/dev/null | awk '{print $$1}'); \
	echo "→ Backend  http://$${LAN_IP:-localhost}:8080"; \
	echo "→ Frontend http://$${LAN_IP:-localhost}:5173"; \
	trap 'echo "→ 停止子进程..."; kill -TERM -$$BPID -$$FPID 2>/dev/null; wait 2>/dev/null; rm -f .dev/backend.pid .dev/frontend.pid; echo "→ 已停止"' INT TERM EXIT; \
		set -m; \
		(cd backend && NA_ADMIN_DEFAULT_PASSWORD=12345678 NA_DATA_DIR=../data go run ./cmd/server) & BPID=$$!; \
		echo $$BPID > .dev/backend.pid; \
		(cd frontend && pnpm run dev) & FPID=$$!; \
		echo $$FPID > .dev/frontend.pid; \
		echo "  Backend  PID: $$BPID"; \
		echo "  Frontend PID: $$FPID"; \
		wait

dev-backend: ## 启动后端开发服务器 (:8080)
	@echo "→ Backend  http://localhost:8080"
	@cd backend && NA_ADMIN_DEFAULT_PASSWORD=12345678 NA_DATA_DIR=../data go run ./cmd/server

dev-frontend: ## 启动前端开发服务器 (:5173)
	@echo "→ Frontend http://localhost:5173"
	@cd frontend && pnpm run dev

dev-cli: ## 编译并运行 CLI（参数通过 ARGS 传递，如 make dev-cli ARGS="task list --family-id xxx"）
	@cd cli && go run . $(ARGS)

# ─── Build ────────────────────────────────────────────────────────

build: build-frontend build-backend build-cli ## 构建全部（前端→后端→CLI，后端嵌入前端制品）

build-backend: ## 编译后端二进制 → backend/na-server（需先 build-frontend）
	@echo "Building backend..."
	@mkdir -p backend/internal/webui/dist
	@if [ -d frontend/dist ]; then \
		rm -rf backend/internal/webui/dist/* && \
		cp -r frontend/dist/* backend/internal/webui/dist/ && \
		echo "  → frontend/dist → backend/internal/webui/dist (embedded)"; \
	else \
		echo "  ⚠️  frontend/dist not found, building without embedded UI"; \
	fi
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

test: test-backend test-cli ## 运行所有 Go 测试

test-backend: ## 运行 backend 测试（含 pkg/）
	@cd backend && go test ./... -count=1 -short

test-cli: ## 运行 CLI 单元测试
	@cd cli && go test ./... -count=1 -short

test-cli-integration: ## 运行 CLI 集成测试（需先 make db-reset && make dev-backend）
	@curl -sf http://localhost:8080/api/system/status >/dev/null 2>&1 || (echo "❌ 后端未运行，请先执行 make dev-backend" && exit 1)
	@cd cli && go test ./tests/integration/... -count=1 -v -timeout 120s

test-e2e: ## 运行 E2E 浏览器自动化测试（需先 make dev 或单独启动服务）
	@echo "→ Installing E2E dependencies..."
	@cd test && npm install --silent 2>/dev/null || true
	@echo "→ Running Playwright tests (chromium)..."
	@cd test && npx playwright test --project=chromium

test-e2e-headed: ## 运行 E2E 测试（有头浏览器，便于调试）
	@cd test && npm install --silent 2>/dev/null || true
	@cd test && npx playwright test --headed --project=chromium

test-e2e-install: ## 安装 Playwright 浏览器
	@cd test && npm install
	@cd test && npx playwright install chromium

docker-e2e-image: ## 构建通用 E2E 工具镜像（无项目代码）
	docker build -t $(E2E_IMAGE) -f Dockerfile.e2e .

docker-e2e-image-zh: ## 构建大陆镜像源 E2E 工具镜像（无项目代码）
	docker build -t $(E2E_IMAGE_ZH) -f Dockerfile.e2e.zh .

define E2E_HEADED_CMD
set -euo pipefail; \
server_log=/tmp/na-server.log; \
rm -f "$$server_log" && touch "$$server_log"; \
cd frontend && pnpm install --frozen-lockfile && NA_STRIP_TEST_ATTRS=0 pnpm build; \
cd $(E2E_WORKSPACE) && rm -rf backend/internal/webui/dist && mkdir -p backend/internal/webui/dist && cp -r frontend/dist/* backend/internal/webui/dist/; \
cd backend && go mod download; \
(NA_ADMIN_DEFAULT_PASSWORD=12345678 NA_DATA_DIR=../data NA_SYNC_TEMPLATES_ON_STARTUP=false go run ./cmd/server > "$$server_log" 2>&1) & \
server_pid=$$!; \
for i in $$(seq 1 90); do \
	if curl -fsS http://127.0.0.1:8080/api/system/status >/dev/null; then break; fi; \
	if ! kill -0 "$$server_pid" 2>/dev/null; then \
		echo "→ backend exited early, log:"; \
		cat "$$server_log"; \
		exit 1; \
	fi; \
	sleep 1; \
done; \
if ! curl -fsS http://127.0.0.1:8080/api/system/status >/dev/null; then \
	echo "→ backend failed to become ready, log:"; \
	cat "$$server_log"; \
	exit 1; \
fi; \
cd $(E2E_WORKSPACE)/test && npm install && xvfb-run -a npx playwright test --project=chromium --headed --reporter=list
endef

docker-e2e-headed: docker-e2e-image ## 在容器内挂载当前目录并用 xvfb 运行 headed E2E（默认镜像）
	$(E2E_BASE_CMD) $(E2E_IMAGE) bash -lc '$(E2E_HEADED_CMD)'

docker-e2e-headed-zh: docker-e2e-image-zh ## 在容器内挂载当前目录并用 xvfb 运行 headed E2E（大陆镜像源）
	$(E2E_BASE_CMD) $(E2E_IMAGE_ZH) bash -lc '$(E2E_HEADED_CMD)'

# ─── Lint / Vet ───────────────────────────────────────────────────

lint: lint-backend lint-frontend ## 运行所有代码检查

lint-backend: ## Go vet backend + CLI
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

check-plugin-isolation: ## 静态规则：主流程/插件边界隔离检查（禁止结构与具体 kind 泄漏）
	@./scripts/check_plugin_isolation.sh

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
	@rm -rf backend/internal/webui/dist/
	@rm -f backend/*.db backend/*.db-journal backend/*.db-wal backend/*.db-shm
	@echo "→ done"

# ─── Dependencies ─────────────────────────────────────────────────

deps: ## 安装所有依赖
	@echo "→ Go modules (backend)..."
	@cd backend && go mod tidy
	@echo "→ Go modules (CLI)..."
	@cd cli && go mod tidy
	@echo "→ pnpm (frontend)..."
	@cd frontend && pnpm install
	@echo "→ all dependencies ready"

# ─── CI ───────────────────────────────────────────────────────────

ci: deps check-contracts check-plugin-isolation lint test build ## CI 完整流水线
	@echo "→ CI passed ✅"

# ─── Docker ───────────────────────────────────────────────────────

docker-build: ## 构建 Docker 镜像
	docker build -t now-and-again .
	docker build -t now-and-again-cli -f cli/Dockerfile .

docker-build-zh: ## 构建 Docker 镜像
	docker build -t now-and-again . -f Dockerfile.zh
	docker build -t now-and-again-cli -f cli/Dockerfile.zh .

docker-up: ## 启动 docker-compose
	docker compose up -d

docker-down: ## 停止 docker-compose
	docker compose down

docker-logs: ## 查看 docker-compose 日志
	docker compose logs -f
