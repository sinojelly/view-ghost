package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"mime"
)

func main() {
    // 强制添加 MIME 类型，防止 Windows 注册表干扰
    mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
    mime.AddExtensionType(".md", "text/markdown; charset=utf-8")
    mime.AddExtensionType(".html", "text/html; charset=utf-8")

	// 获取当前执行目录的绝对路径，避免 Windows 相对路径坑
	executablePath, _ := os.Getwd()
	port := "8080"

	// 自动生成基础文件
	ensureFile("README.md", "# 我的文档库\n\n如果看到这个页面，说明服务已启动。")
	ensureFile("index.html", defaultIndexHTML)

	// 路由设置
	// 1. 侧边栏动态生成 (增加日志)
	http.HandleFunc("/_sidebar.md", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("--> 收到侧边栏请求")
		sidebar := generateSidebar(executablePath)
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		fmt.Fprint(w, sidebar)
		fmt.Println("--- 侧边栏生成完毕 ---")
	})

	// 2. 静态文件服务
	fs := http.FileServer(http.Dir(executablePath))
	http.Handle("/", fs)

	// 打印 IP 方便手机访问
	fmt.Printf("服务运行目录: %s\n", executablePath)
	fmt.Printf("本地访问: http://localhost:%s\n", port)
	fmt.Printf("手机访问: http://%s:%s\n", getLocalIP(), port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("启动失败: %v\n", err)
	}
}

func generateSidebar(root string) string {
	var sb strings.Builder
	sb.WriteString("* [🏠 首页](README.md)\n")

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == root {
			return nil
		}

		name := info.Name()
		// 排除非 MD 文件和配置文件
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || 
		   name == "main.go" || name == "index.html" || name == "README.md" {
			if info.IsDir() { return filepath.SkipDir }
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		webPath := filepath.ToSlash(rel)
		
		// 核心修复：Docsify 的缩进必须是 2 个空格的倍数
		// 我们根据路径中的 "/" 数量来决定缩进层级
		depth := strings.Count(webPath, "/")
		indent := strings.Repeat("  ", depth)

		if info.IsDir() {
			// 文件夹：显示为加粗文本
			sb.WriteString(fmt.Sprintf("%s* **%s**\n", indent, name))
		} else if filepath.Ext(path) == ".md" {
			// 文件：必须是 [显示名](相对路径)
			fileName := strings.TrimSuffix(name, ".md")
			sb.WriteString(fmt.Sprintf("%s* [%s](%s)\n", indent, fileName, webPath))
		}
		return nil
	})
	return sb.String()
}

func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}

func ensureFile(name, content string) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		_ = os.WriteFile(name, []byte(content), 0644)
	}
}

const defaultIndexHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <link rel="stylesheet" href="https://lf3-cdn-tos.bytecdntp.com/cdn/expire-1-M/docsify/4.12.2/themes/vue.css">
</head>
<body>
  <div id="app">正在扫描本地 Markdown 文件...</div>
	<script>
	  window.$docsify = {
		name: '我的文档',
		loadSidebar: true,
		// 关键配置 1：让所有子目录都共用根目录生成的那个 _sidebar.md
		alias: {
		  '/.*/_sidebar.md': '/_sidebar.md'
		},
		// 关键配置 2：如果找不到文件，不要一直转圈，显示 404
		notFoundPage: true,
		// 界面优化
		auto2top: true,
		subMaxLevel: 2,
	  }

	  // 调试助手：如果 5 秒还没加载出来，弹出提示
	  setTimeout(() => {
		if (document.querySelector('#app').innerText.includes('正在扫描')) {
		   document.querySelector('#app').innerHTML = 
		   '<h1>加载超时</h1><p>请检查控制台 (F12) 报错，或确保目录下有 README.md</p>';
		}
	  }, 5000);
	</script>
  <script src="https://lf3-cdn-tos.bytecdntp.com/cdn/expire-1-M/docsify/4.12.2/docsify.min.js"></script>
</body>
</html>`