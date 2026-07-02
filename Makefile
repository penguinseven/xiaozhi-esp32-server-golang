# xiaozhi-esp32-server-golang Makefile
# 本地开发一键操作：依赖安装 → 编译 → 运行
# ==============================================

# ─── 变量定义 ──────────────────────────────────

BINARY           ?= xiaozhi-server
GO               ?= go
CGO_ENABLED      ?= 1

# 配置文件路径
CONFIG           ?= config/config.yaml
MANAGER_CONFIG   ?= build/common/manager.json

# 管理后台模块
MANAGER_DIR       = manager/backend
FRONTEND_DIR      = manager/frontend

# 包路径
CMD_SERVER        = ./cmd/server
CMD_MANAGER       = $(MANAGER_DIR)
CMD_MQTT         ?= ./cmd/mqtt

# 构建标签
TAGS_MANAGER     ?= manager
TAGS_EMBED_UI    ?= embed_ui
TAGS_ASR         ?= asr_server

# 编译选项
LDFLAGS          ?= -s -w

# 颜色输出
RESET  := \033[0m
BOLD   := \033[1m
GREEN  := \033[32m
YELLOW := \033[33m
CYAN   := \033[36m

define info
	@printf "$(BOLD)$(GREEN)[INFO]$(RESET) $(1)\n"
endef

define warn
	@printf "$(BOLD)$(YELLOW)[WARN]$(RESET) $(1)\n"
endef

# ─── 帮助 ──────────────────────────────────────

.PHONY: help
help:
	@echo "$(BOLD)╔════════════════════════════════════════════════╗$(RESET)"
	@echo "$(BOLD)║    xiaozhi-esp32-server-golang 开发命令       ║$(RESET)"
	@echo "$(BOLD)╚════════════════════════════════════════════════╝$(RESET)"
	@echo ""
	@echo "$(BOLD)━━━ 开发环境 ━━━$(RESET)"
	@echo "  $(CYAN)make setup$(RESET)          安装全部依赖（brew + go mod tidy）"
	@echo "  $(CYAN)make deps$(RESET)           仅下载 Go 依赖（go mod tidy）"
	@echo "  $(CYAN)make submodule$(RESET)      拉取子模块（asr_server）"
	@echo ""
	@echo "$(BOLD)━━━ 编译 ━━━$(RESET)"
	@echo "  $(CYAN)make build$(RESET)          编译主程序（不带管理后台）"
	@echo "  $(CYAN)make build-manager$(RESET)  编译管理后台后端"
	@echo "  $(CYAN)make build-aio$(RESET)      编译一体化包（主程序+管理后台+前端嵌入）"
	@echo "  $(CYAN)make build-all$(RESET)      编译所有模块"
	@echo ""
	@echo "$(BOLD)━━━ 运行（单命令）━━━$(RESET)"
	@echo "  $(CYAN)make run$(RESET)            启动主程序（不带管理后台）"
	@echo "  $(CYAN)make run-with-manager$(RESET)  启动主程序（带管理后台内嵌）"
	@echo ""
	@echo "$(BOLD)━━━ 运行（分离部署，推荐开发）━━━$(RESET)"
	@echo "  $(CYAN)make dev-frontend$(RESET)   启动前端 DevServer（热更新）"
	@echo "  $(CYAN)make dev-manager$(RESET)    启动管理后台后端（独立进程）"
	@echo "  $(CYAN)make dev-server$(RESET)     启动主程序（独立进程）"
	@echo "  $(CYAN)make dev$(RESET)            一键启动全部三个进程（前后端+主程序）"
	@echo ""
	@echo "$(BOLD)━━━ 前端 ━━━$(RESET)"
	@echo "  $(CYAN)make frontend-install$(RESET)  安装前端依赖"
	@echo "  $(CYAN)make frontend-build$(RESET)    构建前端生产包"
	@echo "  $(CYAN)make frontend-copy$(RESET)     复制前端输出到 embed 目录"
	@echo ""
	@echo "$(BOLD)━━━ 清理 ━━━$(RESET)"
	@echo "  $(CYAN)make clean$(RESET)          清除编译产物"
	@echo "  $(CYAN)make distclean$(RESET)      彻底清理（含前端 node_modules）"
	@echo ""

# ─── 开发环境 ──────────────────────────────────

.PHONY: setup deps submodule

setup: deps submodule
	$(call info,安装 macOS 系统依赖...)
	@brew list opus 2>/dev/null || brew install opus
	@brew list opusfile 2>/dev/null || brew install opusfile
	@brew list pkg-config 2>/dev/null || brew install pkg-config
	$(call info,检查 OnnxRuntime 头文件...)
	@ls /usr/local/include/onnxruntime_c_api.h 2>/dev/null || $(call warn,缺少 onnxruntime_c_api.h，请运行: make setup-onnx)

deps:
	$(call info,下载 Go 依赖（主模块）...)
	$(GO) mod tidy
	$(call info,下载 Go 依赖（管理后台模块）...)
	cd $(MANAGER_DIR) && $(GO) mod tidy

submodule:
	$(call info,拉取子模块...)
	git submodule update --init --recursive

setup-onnx:
	$(call info,下载并安装 OnnxRuntime 1.21.0（Intel Mac）...)
	cd /tmp && \
	curl -L -o onnxruntime-osx-x86_64-1.21.0.tgz \
	  "https://github.com/microsoft/onnxruntime/releases/download/v1.21.0/onnxruntime-osx-x86_64-1.21.0.tgz" && \
	tar -xzf onnxruntime-osx-x86_64-1.21.0.tgz && \
	sudo cp onnxruntime-osx-x86_64-1.21.0/include/onnxruntime_c_api.h /usr/local/include/ && \
	sudo cp onnxruntime-osx-x86_64-1.21.0/include/onnxruntime_cxx_api.h /usr/local/include/ && \
	sudo cp onnxruntime-osx-x86_64-1.21.0/include/onnxruntime_cxx_inline.h /usr/local/include/ && \
	sudo cp onnxruntime-osx-x86_64-1.21.0/lib/libonnxruntime*.dylib /usr/local/lib/ && \
	$(call info,OnnxRuntime 安装完成)

# ─── 编译 ──────────────────────────────────────

.PHONY: build build-manager build-aio build-all

build: export CGO_ENABLED=$(CGO_ENABLED)
build:
	$(call info,编译主程序（不带管理后台）...)
	$(GO) build -o $(BINARY) $(CMD_SERVER)

build-manager: export CGO_ENABLED=$(CGO_ENABLED)
build-manager:
	$(call info,编译管理后台后端...)
	cd $(MANAGER_DIR) && $(GO) build -o main .

build-aio: frontend-copy
	$(call info,编译一体化包（主程序+管理后台+前端嵌入）...)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build \
	  -tags "$(TAGS_MANAGER) $(TAGS_EMBED_UI)" \
	  -ldflags "$(LDFLAGS)" \
	  -o $(BINARY) $(CMD_SERVER)

build-all: build build-manager
	$(call info,全部编译完成)

# ─── 运行（单命令）─────────────────────────────

.PHONY: run run-with-manager

run: export CGO_ENABLED=$(CGO_ENABLED)
run:
	$(call info,启动主程序...)
	$(GO) run $(CMD_SERVER) -c $(CONFIG)

run-with-manager: frontend-copy
	$(call info,启动主程序（带管理后台）...)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) run \
	  -tags "$(TAGS_MANAGER) $(TAGS_EMBED_UI)" \
	  $(CMD_SERVER) \
	  -c $(CONFIG) \
	  --manager-enable \
	  --manager-config $(MANAGER_CONFIG)

# ─── 运行（分离部署，推荐开发调试用）──────────

.PHONY: dev-frontend dev-manager dev-server dev

dev-frontend:
	$(call info,启动前端 DevServer（http://localhost:3000）...)
	cd $(FRONTEND_DIR) && npm run dev

dev-manager:
	$(call info,启动管理后台后端（http://localhost:8080）...)
	cd $(MANAGER_DIR) && $(GO) run . -c config/config.json

dev-server:
	$(call info,启动主程序（独立进程）...)
	$(GO) run $(CMD_SERVER) -c $(CONFIG) --manager-enable

# 一键启动全部三个进程（需要三个终端窗口）
dev:
	@echo "$(BOLD)请分别在三个终端中执行：$(RESET)"
	@echo ""
	@echo "  $(CYAN)终端 1：$(RESET) make dev-frontend   → http://localhost:3000"
	@echo "  $(CYAN)终端 2：$(RESET) make dev-manager    → http://localhost:8080"
	@echo "  $(CYAN)终端 3：$(RESET) make dev-server     → ws://localhost:8989"
	@echo ""
	@echo "启动顺序建议：终端2(manager) → 终端3(server) → 终端1(frontend)"

# ─── 前端 ──────────────────────────────────────

.PHONY: frontend-install frontend-build frontend-copy

frontend-install:
	$(call info,安装前端依赖...)
	cd $(FRONTEND_DIR) && npm install

frontend-build:
	$(call info,构建前端生产包...)
	cd $(FRONTEND_DIR) && npm run build

frontend-copy: frontend-build
	$(call info,复制前端输出到 embed 目录...)
	@mkdir -p $(MANAGER_DIR)/static/dist
	@cp -r $(FRONTEND_DIR)/dist/* $(MANAGER_DIR)/static/dist/

# ─── MQTT 服务 ─────────────────────────────────

.PHONY: run-mqtt build-mqtt

run-mqtt:
	$(call info,启动独立 MQTT 服务...)
	$(GO) run $(CMD_MQTT) -c $(CONFIG)

build-mqtt:
	$(call info,编译独立 MQTT 服务...)
	$(GO) build -o xiaozhi-mqtt $(CMD_MQTT)

# ─── 清理 ──────────────────────────────────────

.PHONY: clean distclean

clean:
	$(call info,清除编译产物...)
	rm -f $(BINARY)
	rm -f xiaozhi-mqtt
	rm -f $(MANAGER_DIR)/main
	rm -rf $(MANAGER_DIR)/static/dist/
	$(GO) clean -cache

distclean: clean
	$(call info,彻底清理（含前端 node_modules）...)
	rm -rf $(FRONTEND_DIR)/node_modules
	rm -rf $(FRONTEND_DIR)/dist
