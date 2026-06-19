# Unified AgentVoir image: gateway + registry-api + token-accounting in one container.
# Published to ghcr.io/<owner>/agent-voir:<tag> on GitHub Release.
# Build context: repository root.

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

COPY packages/auth-go/go.mod packages/auth-go/go.sum ./packages/auth-go/
COPY packages/auth-go/ ./packages/auth-go/
WORKDIR /src/packages/auth-go
RUN go mod download

WORKDIR /src
COPY apps/gateway/go.mod apps/gateway/go.sum ./apps/gateway/
WORKDIR /src/apps/gateway
RUN go mod download
COPY apps/gateway/ ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /out/agentvoir-gateway ./cmd/gateway

WORKDIR /src
COPY apps/registry-api/go.mod apps/registry-api/go.sum ./apps/registry-api/
WORKDIR /src/apps/registry-api
RUN go mod download
COPY apps/registry-api/ ./
COPY db/migrations/postgres /db/migrations/postgres
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /out/agentvoir-registry-api ./cmd/registry-api

WORKDIR /src
COPY services/token-accounting/go.mod services/token-accounting/go.sum ./services/token-accounting/
WORKDIR /src/services/token-accounting
RUN go mod download
COPY services/token-accounting/ ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /out/agentvoir-token-accounting ./cmd/token-accounting

FROM alpine:3.20
RUN apk add --no-cache tini wget
RUN adduser -D -g '' agentvoir

COPY --from=build /out/agentvoir-gateway /usr/local/bin/agentvoir-gateway
COPY --from=build /out/agentvoir-registry-api /usr/local/bin/agentvoir-registry-api
COPY --from=build /out/agentvoir-token-accounting /usr/local/bin/agentvoir-token-accounting
COPY --from=build /db/migrations/postgres /db/migrations/postgres
COPY deployments/docker/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER agentvoir
ENV MIGRATIONS_DIR=/db/migrations/postgres
EXPOSE 8080 8081 8082
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/entrypoint.sh"]
