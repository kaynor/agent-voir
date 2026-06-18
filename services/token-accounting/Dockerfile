FROM golang:1.25-alpine AS build
WORKDIR /src
COPY services/token-accounting/go.mod services/token-accounting/go.sum ./
RUN go mod download
COPY services/token-accounting/ ./
RUN go build -o /out/agentvoir-token-accounting ./cmd/token-accounting

FROM alpine:3.20
RUN adduser -D -g '' agentvoir
USER agentvoir
COPY --from=build /out/agentvoir-token-accounting /usr/local/bin/agentvoir-token-accounting
EXPOSE 8082
ENTRYPOINT ["agentvoir-token-accounting"]
