# ViewGhost 👻

**ViewGhost** 是一款极致轻量、单文件、跨平台的 Markdown 文档实时预览与文件共享引擎。通过 Go 内嵌静态资源技术，它能将任何本地目录瞬间转化为支持多端访问的精美文档库。

---

## ✨ 核心功能

* **即插即用**：单二进制文件运行，无需预装 Node.js 或任何 Web 服务器。
* **动态侧边栏**：实时扫描当前目录，自动生成 Markdown 目录树，文件变动即刻同步。
* **双能合一**：既是 **Markdown 文档站**，也是 **HTTP 文件服务器**（支持目录浏览与文件下载）。
* **智能首页**：自动识别 `README.md`，若缺失则提供优雅的虚拟介绍页面。
* **多端适配**：适配 PC、手机和平板，支持局域网内所有设备访问。
* **完全离线**：所有渲染资源（JS/CSS/Fonts）均内嵌于程序中，无网环境也能流畅预览。

---

## 🚀 快速开始

### 方案 A：直接运行（推荐）
1.  **下载**：获取对应平台的 `viewghost.exe` (Windows) 或 `viewghost` (Linux)。
2.  **执行**：在命令行中进入你的文档目录，直接输入程序名运行：
    ```bash
    cd D:\MyDocs
    viewghost.exe
    ```
3.  **访问**：浏览器打开控制台显示的地址（默认 `http://localhost:8080`）。

### 方案 B：使用 Docker 部署
通过 Docker，你可以快速在 Linux 服务器上部署文档服务：
```bash
# 1. 运行容器，将文档目录映射到 /app，共享文件目录映射到 /app/shared-files
docker run -d \
  -p 8080:8080 \
  -v /your/docs:/app \
  -v /your/files:/app/shared-files \
  --name viewghost viewghost:latest
```

---

## ⚙️ 进阶配置：`viewghost.config`

在你执行命令的当前文档目录下创建 viewghost.config，程序启动时会自动加载该目录下的个性化配置。
配置采用 `Key=Value` 格式，其他行作为忽略路径：

| 配置项 | 说明 | 示例 |
| :--- | :--- | :--- |
| **PORT** | 定义 Web 服务监听端口（默认 8080） | `PORT=9000` |
| **FILE_PATH** | 指定文件服务器的物理路径，配置后侧边栏将出现“文件共享”入口 | `FILE_PATH=C:\Downloads`，或者是相对于当前文档目录的相对路径。 |
| **(路径/文件名)** | 每一行写一个需要隐藏的目录或文件名（类似 .gitignore） | `.git`, `node_modules` |

**示例配置文件内容：**
```text
PORT=8080
FILE_PATH=./shared-files

# 以下为忽略名单
.git
assets
main.go
viewghost.exe
```

---

## 👨‍💻 开发者指南

ViewGhost 采用 **Go + Docsify** 架构。静态资源（HTML/JS/CSS）通过 Go 的 `embed` 特性直接打包进二进制文件中。

### 1. 开发环境初始化
克隆代码到本地后，执行以下命令初始化环境：
```bash
# 初始化模块
go mod init viewghost
# 整理并下载依赖
go mod tidy
```

### 2. 项目编译
项目支持一键编译为单执行文件。
* **使用 build.bat**：直接运行根目录下的 `build.bat`。
* **手动编译**：
    ```bash
    # 编译全功能版本（包含内嵌资源）
    go build -o viewghost.exe .
    ```

### 3. 便携性说明
编译生成的 `viewghost.exe` 是**完全自包含**的。你可以将它拷贝到任何没有 Go 环境的电脑上，只要将其路径加入系统的 `PATH` 环境变量，即可在任意 Markdown 文档目录下通过命令行一键启动文档预览服务。

---

## 📖 注意事项
* **文件服务器**：文件服务器的 Web 路由固定为 `/fileserver/`，侧边栏会自动集成。
* **首页加载**：如果当前目录没有 `README.md`，ViewGhost 会动态生成一个虚拟主页，确保侧边栏功能依然可用。用户可以在文档目录写入一个 README.md 来定制该页面内容。
* **资源引用**：Markdown 内部图片建议使用相对路径引用。

---

**ViewGhost** —— 让你的本地笔记与资源，从此如影随形。