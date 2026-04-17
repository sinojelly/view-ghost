要通过 Docker 方式将 **ViewGhost** 部署到 Linux 服务器，由于我们的程序已经是单文件二进制（内嵌了所有静态资源），镜像构建会非常简单且体积极小。

### 1. 准备部署文件
在你的项目根目录下（即包含 `main.go`, `fileserver.go`, `index.html`, `assets/` 等文件的目录），创建一个名为 `Dockerfile` 的文件：

**Dockerfile**
```dockerfile
# 第一阶段：编译阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制所有源码和资源
COPY . .

# 初始化模块并编译 (由于是克隆的源码，确保这里能跑通)
RUN go mod init viewghost || true
RUN go mod tidy
# 编译生成名为 viewghost 的 Linux 二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o viewghost .

# 第二阶段：运行阶段
FROM alpine:latest

WORKDIR /app

# 从编译阶段拷贝二进制文件
COPY --from=builder /app/viewghost .
# 拷贝示例配置文件（或者由用户挂载）
COPY --from=builder /app/viewghost.config.example ./viewghost.config

# 暴露端口 (根据你的默认配置，通常是 8080)
EXPOSE 8080

# 运行程序
ENTRYPOINT ["./viewghost"]
```

---

### 2. 编写 `docker-compose.yml` (推荐方式)
使用 Docker Compose 可以更方便地挂载你的 **Markdown 目录** 和 **共享文件目录**。

**docker-compose.yml**
```yaml
version: '3.8'
services:
  viewghost:
    build: .
    container_name: viewghost-app
    ports:
      - "8080:8080"  # 宿主机端口:容器端口
    volumes:
      # 挂载文档目录到容器的工作目录（让程序扫描）
      - "/path/to/your/markdown/docs:/app"
      # 挂载共享文件目录
      - "/path/to/your/shared-files:/app/shared-files"
      # 挂载自定义配置文件
      - "./viewghost.config:/app/viewghost.config"
    restart: always
```

---

### 3. 部署步骤

1. **将源码上传到 Linux 服务器**。
2. **构建镜像**：
   ```bash
   docker build -t viewghost:v1 .
   ```
3. **启动容器**：
   ```bash
   docker-compose up -d
   ```

---

### 4. 针对 Linux 环境的 `viewghost.config` 调整
在 Linux 上部署时，路径格式与 Windows 不同，请确保你的 `viewghost.config` 使用 Linux 路径风格：

```text
PORT=8080

# 注意：这里要写容器内部的路径
# 因为我们在 docker-compose 里把共享目录挂载到了 /app/shared-files
FILE_PATH=/app/shared-files

# 忽略名单
.git
node_modules
viewghost
```

---

### 5. 为什么这种部署方案最简单？

* **体积极小**：基于 `alpine` 的镜像通常只有 20MB 左右。
* **无需环境**：Linux 服务器上不需要安装 Go，所有的依赖都已在 Docker 编译阶段处理完毕。
* **易于迁移**：只要有 Docker，你的整个文档系统（程序 + 资源 + 文档）可以通过一个 `docker-compose.yml` 文件在任何服务器上瞬间复刻。

**提示**：由于我们在 `main.go` 中使用了 `os.Getwd()` 来确定工作目录，在 Docker 中运行程序时，它会自动扫描挂载到 `/app` 下的所有内容。