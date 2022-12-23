FROM golang:1.16.15 AS builder

WORKDIR /build
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags "-extldflags -static" -o webapp-go main.go

FROM alpine:3.16.3

LABEL maintainer="nekoimi <nekoimime@gmail.com>"

COPY --from=builder /build/webapp-go /go/bin/webapp-go

ENV TZ=Asia/Shanghai
ENV PORT=80

WORKDIR /workspace
RUN mkdir -p /public

EXPOSE 80

CMD ["/go/bin/webapp-go"]
