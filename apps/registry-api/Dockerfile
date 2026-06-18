FROM golang:1.25-alpine AS build
WORKDIR /src
COPY apps/registry-api/go.mod apps/registry-api/go.sum ./
RUN go mod download
COPY apps/registry-api/ ./
COPY db/migrations/postgres /db/migrations/postgres
RUN go build -o /out/agentvoir-registry-api ./cmd/registry-api

FROM alpine:3.20
RUN adduser -D -g '' agentvoir
USER agentvoir
ENV MIGRATIONS_DIR=/db/migrations/postgres
COPY --from=build /out/agentvoir-registry-api /usr/local/bin/agentvoir-registry-api
COPY --from=build /db/migrations/postgres /db/migrations/postgres
EXPOSE 8081
ENTRYPOINT ["agentvoir-registry-api"]
