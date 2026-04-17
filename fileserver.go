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
	
	// 修复：计算准确的“上一级”绝对路径
    // 确保 route 格式为 /fileserver
    cleanRoute := "/" + strings.Trim(route, "/")
    
    // 如果 subPath 已经是在根目录了，就不显示返回上一级
    if subPath != "" && subPath != "/" && subPath != "." {
        // 使用 filepath.Dir 获取上一级目录，并转为 URL 格式
        parentPath := filepath.ToSlash(filepath.Dir(strings.TrimSuffix(subPath, "/")))
        if parentPath == "." {
            parentPath = ""
        }
        // 构造绝对路径： /fileserver/ + parentPath
        parentURL := cleanRoute + "/" + strings.TrimPrefix(parentPath, "/")
        fmt.Fprintf(w, "<div class='file-item'><a href='%s'>⬆️ 返回上一级</a></div>", parentURL)
    }

	// ... 后续遍历文件的 a 标签也建议使用绝对路径防止丢失上下文 ...
    for _, f := range files {
        name := f.Name()
        linkName := name
        if f.IsDir() {
            name += "/"
            linkName += "/"
        }
        // 使用绝对路径跳转： /fileserver/subpath/name
        fullLink := filepath.ToSlash(filepath.Join(r.URL.Path, linkName))
        fmt.Fprintf(w, "<div class='file-item'><a href='%s'>%s %s</a></div>", fullLink, getIcon(f), name)
    }
	fmt.Fprintf(w, "<hr><p style='font-size:0.8em;color:#888;'>ViewGhost File Server Mode</p></body></html>")
}

func getIcon(f os.DirEntry) string {
	if f.IsDir() { return "📁" }
	return "📄"
}