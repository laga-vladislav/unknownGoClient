FROM golang:1.24.4-alpine AS go

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bin/server.bin ./cmd/server/

FROM alpine:latest
WORKDIR /app
COPY --from=go /app/bin/server.bin .
COPY --from=go /app/.env .

EXPOSE ${PORT:-7342}

CMD [ "./server.bin" ]