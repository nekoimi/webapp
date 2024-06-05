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

COPY --from=builder /build/webapp /webapp

ENV TZ=Asia/Shanghai

WORKDIR /workspace

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
