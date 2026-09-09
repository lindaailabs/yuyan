# 多阶段构建：构建期 golang:1.25-alpine，运行期 alpine（guide §3.3 单体部署）。
# 注意：构建上下文必须是仓库根目录（需要 packages/protocol 本地 replace）：
#   docker build -f apps/server/Dockerfile -t yuyan-server .
FROM golang:1.25-alpine AS builder

# 构建环境网络受限时使用国内代理（仅构建期，不影响产物）。
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /src

# 先拷依赖清单，利用层缓存（go.mod 的 replace 指向 ../../packages/protocol）。
COPY packages/protocol/go.mod ./packages/protocol/go.mod
COPY apps/server/go.mod apps/server/go.sum ./apps/server/
RUN cd apps/server && go mod download

COPY packages/protocol ./packages/protocol
COPY apps/server ./apps/server
RUN cd apps/server && CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/server /usr/local/bin/server
# 生产配置：密钥经 ${ENV_VAR} 引用环境变量，不落明文；线上启动指定 prod 加载本文件。
COPY --from=builder /src/apps/server/config.prod.yaml /config.prod.yaml

ENV GIN_MODE=release
ENV APP_ENV=prod
ENV CONFIG_FILE=/config.prod.yaml
EXPOSE 8080
ENTRYPOINT ["server"]
