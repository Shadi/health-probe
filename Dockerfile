FROM golang:1.21-alpine AS builder

WORKDIR /app 

COPY . .

RUN apk --no-cache add ca-certificates

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" .

FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/health-probe /usr/bin/

EXPOSE 9100 8080

ENTRYPOINT ["health-probe"]
