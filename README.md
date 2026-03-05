# 📦 Docker Webapp

一个简化前端 Web 项目部署的 Docker 运行环境，通过环境变量动态替换配置，实现**一次构建，处处部署**。

[![Docker Image Size](https://img.shields.io/docker/image-size/nekoimi/webapp/latest)](https://hub.docker.com/r/nekoimi/webapp)
[![Docker Pulls](https://img.shields.io/docker/pulls/nekoimi/webapp)](https://hub.docker.com/r/nekoimi/webapp)
[![GitHub](https://img.shields.io/github/stars/nekoimi/webapp)](https://github.com/nekoimi/webapp)

## ✨ 功能特性

- 🚀 **环境变量注入** - 启动时自动替换前端代码中的环境变量占位符
- 🔧 **零配置部署** - 无需重新构建即可适配不同环境
- 🐳 **Docker 原生支持** - 基于 Nginx Alpine 镜像，轻量高效
- 📁 **多格式支持** - 自动处理 `.html` `.js` `.css` `.json` 文件
- 🛡️ **智能路径处理** - 自动处理斜杠，避免路径拼接问题
- ⚡ **高性能** - 内置 Gzip 压缩和静态资源缓存策略

## 🎯 解决什么问题

前端项目打包后，后端 API 地址、环境变量等配置已硬编码到构建产物中。当需要在不同环境（开发、测试、生产）部署时，通常需要重新构建项目。

**本项目通过环境变量动态替换机制，实现：**

1. 构建时将配置项以变量名形式打包（如 `API_SERVER_URL`）
2. 部署时通过环境变量传入实际值
3. 容器启动时自动替换，无需重新构建

## 📥 快速开始

### 拉取镜像

```bash
docker pull ghcr.io/nekoimi/webapp:latest
```

### 运行示例

```bash
git clone https://github.com/nekoimi/webapp.git
cd webapp
docker-compose up -d
```

访问 http://localhost 查看效果

## 🛠️ 使用方法

### 1. 准备前端项目

在前端代码中使用环境变量名称作为占位符：

```html
<!-- index.html -->
<script>
  const apiUrl = 'API_SERVER_URL';  // 将被替换为实际值
  const appName = 'APP_NAME';       // 将被替换为实际值
</script>
```

### 2. 使用 Docker Compose

```yaml
version: "3.6"
services:
  webapp:
    image: ghcr.io/nekoimi/webapp:latest
    ports:
      - "80:80"
    environment:
      # 以 WEBAPP_ENV. 为前缀设置环境变量
      WEBAPP_ENV.API_SERVER_URL: https://api.example.com
      WEBAPP_ENV.APP_NAME: 我的应用
      WEBAPP_ENV.BASE_URL: /app/
      WEBAPP_ENV.USERNAME: admin
      WEBAPP_ENV.PASSWORD: secret123
    volumes:
      - ./dist:/workspace  # 挂载前端构建产物
```

### 3. 构建自定义镜像

```dockerfile
FROM ghcr.io/nekoimi/webapp:latest

COPY ./dist /workspace
```

```bash
docker build -t my-webapp .
docker run -p 80:80 -e WEBAPP_ENV.API_URL=https://api.example.com my-webapp
```

## ⚙️ 配置说明

### 环境变量

| 变量名 | 说明 | 示例 |
|--------|------|------|
| `WEBAPP_ENV.<KEY>` | 自定义配置项，将被替换到前端代码中 | `WEBAPP_ENV.API_URL=https://api.com` |
| `PORT` | HTTP 服务端口（默认 80） | `PORT=8080` |
| `TZ` | 时区设置（默认 Asia/Shanghai） | `TZ=Asia/Shanghai` |

### 支持的文件类型

以下文件类型会自动进行环境变量替换：

- `.html` - HTML 文件
- `.js` - JavaScript 文件
- `.css` - 样式文件
- `.json` - JSON 配置文件

### 路径处理规则

工具会自动处理斜杠，避免以下问题：

- `//api` → `/api`（自动去重）
- `api/` + `/endpoint` → `api/endpoint`（智能拼接）

## 📁 项目结构

```
.
├── main.go              # Go 部署工具源码
├── Dockerfile           # 镜像构建配置
├── docker-compose.yaml  # 示例编排文件
├── conf/                # Nginx 配置
│   ├── nginx.conf       # Nginx 主配置
│   └── default.conf.tpl # 站点配置模板
├── example/             # 示例前端项目
│   ├── index.html
│   ├── css/
│   └── js/
└── 99-run-webapp.sh     # 容器启动脚本
```

## 🔧 工作原理

1. **启动阶段**：容器执行 `99-run-webapp.sh` 脚本
2. **环境加载**：`webapp` 程序读取 `WEBAPP_ENV.*` 前缀的环境变量
3. **文件复制**：将 `/workspace` 目录内容复制到 Nginx 根目录
4. **变量替换**：扫描 `.html` `.js` `.css` `.json` 文件，替换占位符
5. **权限设置**：设置文件权限，启动 Nginx 服务

## 🌟 最佳实践

### 前端项目配置

建议使用统一的环境变量命名规范：

```javascript
// config.js
const config = {
  apiUrl: 'API_SERVER_URL',
  appName: 'APP_NAME',
  baseUrl: 'BASE_URL',
  version: 'APP_VERSION'
};
```

### CI/CD 集成

```yaml
# .github/workflows/deploy.yml
- name: Deploy to Production
  run: |
    docker run -d \
      -p 80:80 \
      -e WEBAPP_ENV.API_SERVER_URL=${{ secrets.PROD_API_URL }} \
      -e WEBAPP_ENV.APP_NAME=生产环境 \
      -v $(pwd)/dist:/workspace \
      ghcr.io/nekoimi/webapp:latest
```

## 🐛 故障排查

### 环境变量未生效

1. 检查变量名是否以 `WEBAPP_ENV.` 开头
2. 查看容器日志：`docker logs <container_id>`
3. 确认挂载目录正确：`/workspace` 应包含前端文件

### 权限问题

```bash
# 检查文件权限
docker exec <container_id> ls -la /usr/share/nginx/html

# 修复权限
chmod -R 755 ./dist
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

[MIT License](LICENSE)

---

**Made with ❤️ by [nekoimi](https://github.com/nekoimi)**
