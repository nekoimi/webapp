# 📦 docker webapp

一个简化部署的web项目运行环境，web项目docker化，简化不同环境的部署。

前端web项目打包好之后生成的dist中相关的后端API、环境变量配置都已写到相关代码中，
在docker化之后切换环境部署的时候非常不方便，需要重新构建项目生成对应环境的dist，
也许不同环境的唯一区别就是后端服务器的API地址不一样。这里使用配置环境变量的方式，
在项目启动的时候，自动根据环境变量来替换相关的配置，达到一处构建处处部署的目的，
简化项目部署。

[![Docker Image Size (tag)](https://img.shields.io/docker/image-size/nekoimi/webapp/latest)](https://hub.docker.com/r/nekoimi/webapp)
[![Docker Pulls](https://img.shields.io/docker/pulls/nekoimi/webapp)](https://hub.docker.com/r/nekoimi/webapp)

# 📥 Download

### Docker Image

- [DockerHub](https://hub.docker.com/r/nekoimi/webapp)

```bash
docker pull ghcr.io/nekoimi/webapp:latest
```

<br>

# 🛠️ 使用

- 项目打包时，约定配置对应的环境变量名称，将环境变量名称作为配置打包进dist

- 在启动docker image的时候，以`WEBAPP_ENV.`为环境变量前缀设置对应的环境变量值即可

### 例子

生成dist产物例子: [repo](https://github.com/nekoimi/docker-webapp-go.git)

step1:

```bash
git clone https://github.com/nekoimi/docker-webapp-go.git
```

step2:

```base
docker-compose up -d
```

step3:

浏览器打开默认访问链接[http://127.0.0.1](http://127.0.0.1)查看应用
<br>

### Using Docker Image

### docker compose

```bash
version: "3.6"
services:
  test:
    image: ghcr.io/nekoimi/webapp:latest
    ports:
      - "80:80"
    environment:
      WEBAPP_ENV.API_SERVER_URL: http://127.0.0.1/api
      WEBAPP_ENV.APP_NAME: 测试web
      WEBAPP_ENV.BACKGROUND_IMAGE: image.png
      WEBAPP_ENV.BASE_URL: /baseurl/
      WEBAPP_ENV.USERNAME: user001
      WEBAPP_ENV.PASSWORD: user001_pwd
    volumes:
      - ./example:/workspace

```

### 项目构建

``` bash
# Dockerfile
FROM ghcr.io/nekoimi/webapp:latest

COPY /dist    /workspace
```


<br>

# 📄 License

[MIT License](#LICENSE)