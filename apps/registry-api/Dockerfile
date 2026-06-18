FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /out/agentvoir-registry-api ./cmd/registry-api

FROM alpine:3.20
RUN adduser -D -g '' agentvoir
USER agentvoir
COPY --from=build /out/agentvoir-registry-api /usr/local/bin/agentvoir-registry-api
EXPOSE 8081
ENTRYPOINT ["agentvoir-registry-api"]
