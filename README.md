# ViewGhost

**ViewGhost** 是一款极致轻量、零配置的本地 Markdown 文档实时预览引擎。它是 **SyncGhost** 的孪生组件，旨在将 Windows 本地目录瞬间转化为支持多端访问（手机/平板）的精美文档库。

---

## 🚀 核心特性

* **实时感知**：无需手动更新索引，增加、删除、移动 Markdown 文件或目录，网页端刷新即刻同步。
* **双端适配**：针对桌面端与移动端浏览器进行深度优化，支持手机局域网顺畅访问。
* **完全离线**：支持本地静态资源引用，无需连接外网即可享受 GitHub 级别的渲染效果。
* **沉浸式体验**：内置对 Typora 自定义样式的支持，让网页预览与本地编辑体验高度统一。
* **智能过滤**：支持通过 `viewghost.config` 自定义隐藏特定文件夹（如 `.git`, `node_modules`）。

---

## 🛠️ 快速开始

### 1. 环境准备
将以下文件放置在你的工作目录：
* `viewghost.exe` (由 `main.go` 编译生成)
* `index.html` (Docsify 配置文件)
* `assets/` (包含 `vue.css`, `github.css`, `docsify.min.js` 等)

### 2. 运行服务
在目录下执行：
```bash
go run main.go
# 或者直接双击编译好的 exe
```

### 3. 访问方式
* **本机**：浏览器访问 `http://localhost:8080`
* **手机**：确保与电脑处于同一 WiFi，访问控制台输出的 IP 地址，例如 `http://192.168.31.194:8080`

---

## ⚙️ 进阶配置

### 忽略文件清单
在根目录创建 `viewghost.config`，每行一个路径，支持过滤：
```text
.git
node_modules
private_notes
temp_drafts
assets
```

---

## 👨‍💻 开发者信息

ViewGhost 采用 **Go + Docsify** 架构开发，致力于提供最轻量化的文档展示方案。

### 持续维护计划
* [ ] 集成 SyncGhost 的实时文件监听 (fsnotify)，实现无需刷新的自动重载。
* [ ] 增加本地图片上传与管理接口。
* [ ] 支持一键导出为静态 HTML 站点。

---

### 💡 小贴士
如果发现排版异常，请确保你的 Markdown 表格前后留有空行。对于图片显示，建议使用相对路径存储在文档同级或 `images/` 目录下。

---

**ViewGhost** —— 让你的本地笔记从此“动”起来。