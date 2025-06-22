FROM golang:1.20.5 AS builder

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/

WORKDIR /build
COPY . .
RUN go install
RUN go build --ldflags "-extldflags -static" -o webapp main.go

FROM nginx:alpine-slim

LABEL maintainer="nekoimi <nekoimime@gmail.com>"

ENV TZ=Asia/Shanghai
ENV PORT=80

COPY --from=builder /build/webapp                   /bin/webapp
COPY --from=builder /build/99-run-webapp.sh         /docker-entrypoint.d/99-run-webapp.sh
COPY --from=builder /build/conf/nginx.conf          /etc/nginx/nginx.conf
COPY --from=builder /build/conf/default.conf.tpl    /etc/nginx/conf.d/default.conf.tpl

RUN chmod +x /docker-entrypoint.d/99-run-webapp.sh

WORKDIR /workspace
