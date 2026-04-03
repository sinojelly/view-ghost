package main

import (
	"bufio"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 定义全局变量存储配置
var (
	appPort = "8080" // 默认端口
	ignored []string
)

func main() {
	// 1. 初始化配置
	loadAppConfig("viewghost.config")

	// 2. 注册 MIME 类型（防止 Windows 拦截）
	mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")

	executablePath, _ := os.Getwd()

	// 3. 路由设置
	http.HandleFunc("/_sidebar.md", func(w http.ResponseWriter, r *http.Request) {
		sidebar := generateSidebar(executablePath)
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		fmt.Fprint(w, sidebar)
	})

	http.Handle("/", http.FileServer(http.Dir(executablePath)))

	// 4. 启动服务
	fmt.Printf("ViewGhost 已启动！\n")
	fmt.Printf("本地访问: http://localhost:%s\n", appPort)
	fmt.Printf("手机访问: http://%s:%s\n", getLocalIP(), appPort)

	if err := http.ListenAndServe(":"+appPort, nil); err != nil {
		fmt.Printf("端口 %s 被占用或启动失败: %v\n", appPort, err)
	}
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