# Yuyan monorepo 构建入口（Linux/macOS/CI）。
# Windows 无 make 时用等价物：powershell -File scripts/make.ps1 -Target <目标>

GO ?= go
SERVER_DIR := apps/server
PROTOCOL_DIR := packages/protocol

.PHONY: run test-server lint gen-protocol

# 本地直跑 server（前置：docker compose --profile db-only up -d）
run:
	cd $(SERVER_DIR) && $(GO) run ./cmd/server

# 服务端全部测试（含 testcontainers 需要 Docker）
test-server:
	cd $(SERVER_DIR) && $(GO) test ./...

# golangci-lint（配置见根目录 .golangci.yml；安装：go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest）
lint:
	golangci-lint run ./$(SERVER_DIR)/... ./$(PROTOCOL_DIR)/...

# 协议 Schema 校验（占位实现，二期扩展为双端代码生成）
gen-protocol:
	cd $(PROTOCOL_DIR) && $(GO) test ./...
