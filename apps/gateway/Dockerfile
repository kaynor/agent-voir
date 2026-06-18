FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /out/agentvoir-gateway ./cmd/gateway

FROM alpine:3.20
RUN adduser -D -g '' agentvoir
USER agentvoir
COPY --from=build /out/agentvoir-gateway /usr/local/bin/agentvoir-gateway
EXPOSE 8080
ENTRYPOINT ["agentvoir-gateway"]
