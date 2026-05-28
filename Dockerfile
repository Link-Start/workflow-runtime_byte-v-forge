FROM docker.m.daocloud.io/library/node:22-bookworm-slim AS dashboard_remote_builder

WORKDIR /workflow-runtime/webui
COPY common-lib/ui /common-lib/ui
COPY common-lib/proto /common-lib/proto
COPY common-lib/scripts /common-lib/scripts
COPY workflow-runtime/webui ./
RUN npm ci && npm run build

FROM docker.m.daocloud.io/library/golang:1.26-alpine AS builder

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache git ca-certificates

COPY common-lib /common-lib
COPY workflow-runtime/go.mod workflow-runtime/go.sum ./
RUN go mod edit -replace github.com/byte-v-forge/common-lib=/common-lib \
    && go mod download

COPY workflow-runtime .
RUN CGO_ENABLED=0 GOOS=linux go build -o workflow-runtime ./cmd/workflow-runtime

FROM docker.m.daocloud.io/library/alpine:latest

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/workflow-runtime .
COPY --from=dashboard_remote_builder /workflow-runtime/webui/dist /app/dashboard/workflow-runtime
EXPOSE 8080
CMD ["./workflow-runtime"]
