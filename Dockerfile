FROM golang:alpine AS builder
LABEL maintainer="daniele@coolbyte.eu" 
RUN apk --update add \
    go \
    musl-dev
WORKDIR /home-ddns
RUN apk --update add \
    util-linux-dev
# Copy main files
COPY main.go /home-ddns
COPY api/*.go /home-ddns/api/
COPY config/config.go /home-ddns/config/
COPY models/models.go /home-ddns/models/
COPY utils/utils.go /home-ddns/utils/
# Copy Module files
COPY go.mod /home-ddns/ 
COPY go.sum  /home-ddns/
# Download Module dependencies
RUN go mod download 
# Compile the binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" .
RUN adduser \    
    --disabled-password \    
    --shell "/sbin/nologin" \    
    "appuser"

FROM scratch

COPY --from=builder /home-ddns/home-ddns /home-ddns/home-ddns
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

USER appuser:appuser

ENTRYPOINT ["/home-ddns/home-ddns", "-config", "/home-ddns/config.yaml"]

