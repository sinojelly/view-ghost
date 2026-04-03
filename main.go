package main

import (
	"embed"
	"fmt"
    "bufio"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// --- 关键修改：打包静态资源 ---
//假设你的项目结构中有一个 assets 目录存放 css/js，以及同级的 index.html 和 favicon.ico
//go:embed index.html assets/* favicon.ico
var embeddedFiles embed.FS

var (
	appPort = "8080"
	ignored []string
)

func main() {
	// 获取当前工作目录（即用户在命令行执行命令时的位置）
	currentDir, _ := os.Getwd()

	// 1. 加载配置（优先查找当前目录下的 viewghost.config，没有则用默认）
	loadAppConfig(filepath.Join(currentDir, "viewghost.config"))

	mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")

	// 2. 路由处理
	// 动态侧边栏：扫描当前命令行所在的目录
	http.HandleFunc("/_sidebar.md", func(w http.ResponseWriter, r *http.Request) {
		sidebar := generateSidebar(currentDir)
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		fmt.Fprint(w, sidebar)
	})

	// 静态资源处理：
	// 如果请求的是 index.html, assets/*, favicon.ico -> 从【内存(embed)】读取
	// 如果请求的是 .md 或图片 -> 从【本地磁盘】读取
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, "/")
		
		// 优先从嵌入资源中查找（index.html, js, css）
		if reqPath == "" || reqPath == "index.html" || strings.HasPrefix(reqPath, "assets/") || reqPath == "favicon.ico" {
			if reqPath == "" { reqPath = "index.html" }
			data, err := embeddedFiles.ReadFile(reqPath)
			if err == nil {
				// 手动设置 MIME，因为 embed 文件系统有时无法自动识别
				if strings.HasSuffix(reqPath, ".css") { w.Header().Set("Content-Type", "text/css") }
				if strings.HasSuffix(reqPath, ".js") { w.Header().Set("Content-Type", "application/javascript") }
				w.Write(data)
				return
			}
		}

		// 否则从本地磁盘读取（Markdown 文件、图片等）
		http.FileServer(http.Dir(currentDir)).ServeHTTP(w, r)
	})

	fmt.Printf("ViewGhost (Portable Mode) 已启动！\n")
	fmt.Printf("正在工作目录: %s\n", currentDir)
	fmt.Printf("访问地址: http://%s:%s\n", getLocalIP(), appPort)

	http.ListenAndServe(":"+appPort, nil)
}

// 增强型配置加载
func loadAppConfig(filename string) {
	ignored = []string{"node_modules", ".git", "assets", "index.html", "main.go"}
	
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 识别 PORT= 配置
		if strings.HasPrefix(strings.ToUpper(line), "PORT=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 && parts[1] != "" {
				appPort = strings.TrimSpace(parts[1])
			}
			continue
		}

		// 其余作为忽略路径
		cleanPath := filepath.ToSlash(strings.Trim(line, "/"))
		ignored = append(ignored, cleanPath)
	}
}

func generateSidebar(root string) string {
	var sb strings.Builder
	sb.WriteString("* [🏠 首页](README.md)\n")

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == root { return nil }

		rel, _ := filepath.Rel(root, path)
		webPath := filepath.ToSlash(rel)
		name := info.Name()

		// 过滤逻辑
		for _, item := range ignored {
			if strings.HasPrefix(webPath, item) || name == item {
				if info.IsDir() { return filepath.SkipDir }
				return nil
			}
		}

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
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ipStr := ipnet.IP.String()
			if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") {
				return ipStr
			}
		}
	}
	return "127.0.0.1"
}