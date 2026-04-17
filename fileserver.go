package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileServerConfig 存储文件服务器的配置
type FileServerConfig struct {
	RoutePath string // 浏览器访问路径，如 "/download"
	LocalPath string // 本地磁盘路径，如 "D:/SharedFiles"
}

// RegisterFileServer 注册文件服务器路由
func RegisterFileServer(cfg FileServerConfig) {
	if cfg.RoutePath == "" || cfg.LocalPath == "" {
		return
	}

	// 确保路径以 / 开头且不以 / 结尾
	route := "/" + strings.Trim(cfg.RoutePath, "/")

	http.HandleFunc(route+"/", func(w http.ResponseWriter, r *http.Request) {
		// 获取相对于路由的子路径
		subPath := strings.TrimPrefix(r.URL.Path, route)
		fullPath := filepath.Join(cfg.LocalPath, subPath)

		// 检查路径是否存在
		info, err := os.Stat(fullPath)
		if err != nil {
			http.Error(w, "文件或目录不存在", http.StatusNotFound)
			return
		}

		// 如果是目录，生成一个简单的列表页
		if info.IsDir() {
			renderDirectoryListing(w, r, fullPath, route, subPath)
			return
		}

		// 如果是文件，直接触发下载
		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(fullPath))
		http.ServeFile(w, r, fullPath)
	})
}

// 生成简单的目录索引 HTML
func renderDirectoryListing(w http.ResponseWriter, r *http.Request, localDir, route, subPath string) {
	files, err := os.ReadDir(localDir)
	if err != nil {
		http.Error(w, "无法读取目录", 500)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// 在 fileserver.go 的 renderDirectoryListing 中修改 style 标签, 使其背景颜色与 Docsify 保持一致
	fmt.Fprintf(w, "<html><head><style>"+
    "body{font-family:sans-serif; padding:10px 20px; background:#fff; color:#34495e; margin:0;}"+ // 减少 padding 适应 iframe
    "a{color:#42b983; text-decoration:none; font-weight:bold;}"+
    ".file-item{padding:12px; border-bottom:1px solid #f0f0f0; display:flex; align-items:center; transition: background 0.2s;}"+
    ".file-item:hover{background: #f9f9f9;}"+ // 增加悬停反馈
    "</style></head><body>")
	
	fmt.Fprintf(w, "<h2>📂 目录索引: %s</h2><hr>", r.URL.Path)
	
	// 返回上一级
	if subPath != "" && subPath != "/" {
		fmt.Fprintf(w, "<div class='file-item'><a href='..'>⬆️ 返回上一级</a></div>")
	}

	for _, f := range files {
		name := f.Name()
		if f.IsDir() {
			name += "/"
		}
		// 拼接下载链接
		fmt.Fprintf(w, "<div class='file-item'><a href='%s'>%s %s</a></div>", filepath.ToSlash(filepath.Join(r.URL.Path, f.Name())), getIcon(f), name)
	}
	fmt.Fprintf(w, "<hr><p style='font-size:0.8em;color:#888;'>ViewGhost File Server Mode</p></body></html>")
}

func getIcon(f os.DirEntry) string {
	if f.IsDir() { return "📁" }
	return "📄"
}