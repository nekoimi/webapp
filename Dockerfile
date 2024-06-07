FROM golang:1.20.5 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/

WORKDIR /build
COPY . .
RUN go install
RUN go build --ldflags "-extldflags -static" -o webapp main.go

FROM nginx:alpine-slim

LABEL maintainer="nekoimi <nekoimime@gmail.com>"

ENV TZ=Asia/Shanghai

COPY --from=builder /build/webapp               /bin/webapp
COPY --from=builder /build/99-run-webapp.sh     /docker-entrypoint.d/99-run-webapp.sh
COPY --from=builder /build/conf/nginx.conf      /etc/nginx/nginx.conf
COPY --from=builder /build/conf/default.conf    /etc/nginx/conf.d/default.conf

RUN chmod +x /docker-entrypoint.d/99-run-webapp.sh

WORKDIR /workspace
