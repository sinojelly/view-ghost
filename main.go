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
	mime.AddExtensionType(".css", "text/css")

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

	// ":" 代表绑定到所有网卡，包括 127.0.0.1 和 192.168.31.194
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("启动失败: %v\n", err)
	}
}

// 改进的配置加载：处理空格、空行和斜杠统一
func loadConfig(filename string) []string {
	content, err := os.ReadFile(filename)
	if err != nil {
		// 默认忽略项，统一不带前后斜杠
		return []string{"node_modules", ".git", "assets", "main.go", "index.html"}
	}
	
	lines := strings.Split(string(content), "\n")
	var ignored []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 统一转为正斜杠并去掉首尾斜杠，例如 "test/" 变成 "test"
		line = filepath.ToSlash(line)
		line = strings.Trim(line, "/")
		ignored = append(ignored, line)
	}
	return ignored
}

func generateSidebar(root string) string {
	var sb strings.Builder
	sb.WriteString("* [🏠 首页](README.md)\n")

	ignored := loadConfig("viewghost.config")

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == root { return nil }

		// 获取相对路径并统一为正斜杠格式
		rel, _ := filepath.Rel(root, path)
		webPath := filepath.ToSlash(rel)
		name := info.Name()

		// 检查过滤逻辑
		for _, item := range ignored {
			// 匹配逻辑：如果是目录开头匹配，或者文件名完全匹配
			if strings.HasPrefix(webPath, item) || name == item {
				if info.IsDir() {
					return filepath.SkipDir // 关键：跳过整个文件夹的扫描
				}
				return nil
			}
		}

		// 基础隐藏逻辑
		if strings.HasPrefix(name, ".") || name == "README.md" {
			if info.IsDir() { return filepath.SkipDir }
			return nil
		}

		depth := strings.Count(webPath, "/")
		indent := strings.Repeat("  ", depth)

		if info.IsDir() {
			sb.WriteString(fmt.Sprintf("%s* **%s**\n", indent, name))
		} else if filepath.Ext(path) == ".md" {
			sb.WriteString(fmt.Sprintf("%s* [%s](%s)\n", indent, strings.TrimSuffix(name, ".md"), webPath))
		}
		return nil
	})
	return sb.String()
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		// 检查 ip 地址判断是否回环
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				// 优先返回 192.168 或 10.0 开头的局域网常用地址
				if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") {
					return ipStr
				}
			}
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