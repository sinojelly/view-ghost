package main

import (
	"bufio"
	"embed"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// --- 静态资源打包 ---
//go:embed index.html favicon.ico assets/*
var embeddedFiles embed.FS

var (
	appPort = "8080"
	ignored []string
	fileServerPath = "" // 新增：本地文件服务器物理路径
	fileRoute      = "fileserver" // 新增：默认访问路由，并且不支持配置
)

func main() {
	// 获取执行命令时的当前目录
	currentDir, _ := os.Getwd()

	// 1. 加载配置
	loadAppConfig(filepath.Join(currentDir, "viewghost.config"))

	// 2. 注册 MIME 类型
	mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")

    // --- 新增调用：注册文件服务器 ---
	// 注意：FileServer 的路由优先级高于原本的 "/" 拦截，Go 会自动匹配最长的前缀
	if fileServerPath != "" {
		RegisterFileServer(FileServerConfig{
			RoutePath: fileRoute,
			LocalPath: fileServerPath,
		})
	}

    // 3. 核心路由处理
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, "/")
		
		// 如果请求的是 /fileserver 相关的，但由于某种原因漏到了这里
		// 我们可以做一个重定向或者错误处理，防止它加载 index.html
		if strings.HasPrefix(reqPath, fileRoute) {
			// 正常情况下 RegisterFileServer 应该拦截了它
			// 如果走到这里，说明路径可能少了斜杠，补全它
			http.Redirect(w, r, "/"+fileRoute+"/", http.StatusMovedPermanently)
			return
		}

		// --- 1. 动态生成侧边栏 (最高优先级) ---
		if reqPath == "_sidebar.md" {
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			fmt.Fprint(w, generateSidebar(currentDir))
			return
		}

		// --- 2. 根路径 "/"：必须返回 index.html ---
		if reqPath == "" {
			data, err := embeddedFiles.ReadFile("index.html")
			if err != nil {
				http.Error(w, "Internal Server Error: index.html missing", 500)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache") // 强制不缓存
			w.Write(data)
			return
		}

		// --- 3. 智能 README 拦截 (仅当明确请求 README.md 时) ---
		if reqPath == "README.md" {
			localPath := filepath.Join(currentDir, "README.md")
			fileInfo, err := os.Stat(localPath)
			// 如果不存在或为空
			if os.IsNotExist(err) || (err == nil && fileInfo.Size() == 0) {
				w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
				defaultReadme := "# ViewGhost 👻\n\n> 欢迎使用 **ViewGhost**！\n\n当前目录下未检测到 `README.md`，已自动为你开启**浏览模式**。\n\n### 💡 使用提示\n- **查看文档**：请点击左侧菜单栏浏览当前目录下的 Markdown 文件。\n- **自定义首页**：在当前目录下新建一个 `README.md` 文件即可覆盖此内容。"
				fmt.Fprint(w, defaultReadme)
				return
			}
			// 如果存在且有内容，则不 return，继续往下走 FileServer 读取磁盘文件
		}

		// --- 4. 其它静态资源 (assets/*, favicon.ico) ---
		data, err := embeddedFiles.ReadFile(reqPath)
		if err == nil {
			if strings.HasSuffix(reqPath, ".css") { w.Header().Set("Content-Type", "text/css") }
			if strings.HasSuffix(reqPath, ".js") { w.Header().Set("Content-Type", "application/javascript") }
			w.Write(data)
			return
		}

		// --- 5. 最后：读取本地磁盘文件 (MD, 图片等) ---
		http.FileServer(http.Dir(currentDir)).ServeHTTP(w, r)
	})

	fmt.Printf("ViewGhost (Portable Mode) 已启动！\n")
	fmt.Printf("当前工作路径: %s\n", currentDir)
	fmt.Printf("本地访问: http://localhost:%s\n", appPort)
	fmt.Printf("手机访问: http://%s:%s\n", getLocalIP(), appPort)
	if fileServerPath != "" {
		fmt.Printf("文件服务器: http://%s:%s/%s\n", getLocalIP(), appPort, fileRoute)
	}

	if err := http.ListenAndServe(":"+appPort, nil); err != nil {
		fmt.Printf("启动失败: %v\n", err)
	}
}

// 加载配置
func loadAppConfig(filename string) {
	ignored = []string{"node_modules", ".git", "assets", "index.html", "main.go", "viewghost.exe", "viewghost.config"}
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
		if strings.HasPrefix(strings.ToUpper(line), "PORT=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 && parts[1] != "" {
				appPort = strings.TrimSpace(parts[1])
			}
			continue
		}
		// 新增：识别 FILE_PATH= 和 FILE_ROUTE=
		if strings.HasPrefix(strings.ToUpper(line), "FILE_PATH=") {
			fileServerPath = strings.TrimSpace(line[10:])
			continue
		}
		// if strings.HasPrefix(strings.ToUpper(line), "FILE_ROUTE=") {
		//	fileRoute = "fileserver"  //固定，不允许从配置文件配置。否则index.html就不能写死。 strings.TrimSpace(line[11:])
		//	continue
		// }
		cleanPath := filepath.ToSlash(strings.Trim(line, "/"))
		ignored = append(ignored, cleanPath)
	}
}

// 生成侧边栏
func generateSidebar(root string) string {
	var sb strings.Builder
	sb.WriteString("* [🏠 首页](README.md)\n")

	// --- 核心修改：在侧边栏加文件共享入口 ---
	if fileServerPath != "" {
		// --- 修改：指向 Docsify 内部的虚拟路径，而不是后端原始路径 ---
		// 去掉前面的斜杠，确保 Docsify 将其识别为路径切换而非锚点
        sb.WriteString("* [📂 文件共享](fileserver-view)\n")
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == root {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		webPath := filepath.ToSlash(rel)
		name := info.Name()

		// --- 修改：增加对文件服务器路径的过滤 ---
		// 获取配置中的本地共享目录名 (例如 "shared-files")
		sharedDirName := filepath.Base(fileServerPath)

		for _, item := range ignored {
			if strings.HasPrefix(webPath, item) || name == item || name == sharedDirName {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if strings.HasPrefix(name, ".") || name == "README.md" {
			if info.IsDir() {
				return filepath.SkipDir
			}
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

// 获取局域网 IP
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