FROM golang:1.14 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o gateway .

FROM alpine:3.12
RUN apk add --no-cache ca-certificates
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
WORKDIR /app
COPY --from=builder /app/gateway /app
COPY ./config/* ./config/

CMD ["/app/gateway"]