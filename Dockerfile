FROM alpine:3.16.3

LABEL maintainer="nekoimi <nekoimime@gmail.com>"

ENV TZ=Asia/Shanghai
ENV PORY=80
ENV REWRITE=true

RUN mkdir -p /public

WORKDIR /workspace

EXPOSE 80

CMD ["webapp-go"]
