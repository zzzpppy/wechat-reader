package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"compress/gzip"
	"os"
	"path/filepath"
	"wechat-reader/internal/service"
	"wechat-reader/internal/storage"
)

func main() {
	ctx := context.Background()

	// 获取当前工作目录
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化数据库连接
	db, err := storage.NewDatabase(ctx, filepath.Join(pwd, "data.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(ctx)

	// 初始化爬虫服务
	crawler := service.NewCrawler()

	// API 处理函数
	http.HandleFunc("/api/fetch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 解析 JSON 请求体
		var request struct {
			URL string `json:"url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if request.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		articles, err := crawler.FetchArticles(ctx, request.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("articles:", articles)
		// 保存到数据库
		if err := db.SaveArticles(ctx, articles); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 返回结果
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    articles,
			"message": "Articles fetched and saved successfully",
		})
	})

	// 获取文章列表
	http.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		articles, err := db.GetArticles(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    articles,
		})
	})

	// 获取主题列表
	http.HandleFunc("/api/topics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		topics, err := db.GetTopics(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    topics,
		})
	})

	// 添加图片代理接口
	http.HandleFunc("/api/proxy/image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		imageURL := r.URL.Query().Get("url")
		if imageURL == "" {
			http.Error(w, "Image URL is required", http.StatusBadRequest)
			return
		}

		// 创建图片请求
		req, err := http.NewRequest("GET", imageURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 添加必要的请求头
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
		req.Header.Set("Referer", "https://mp.weixin.qq.com/")

		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// 复制响应头
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// 复制图片数据
		io.Copy(w, resp.Body)
	})

	// 修改原有的代理接口
	// 修改代理接口
	http.HandleFunc("/api/proxy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		targetURL := r.URL.Query().Get("url")
		if targetURL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		log.Printf("Proxying request to: %s", targetURL)

		// 创建代理请求
		req, err := http.NewRequest("GET", targetURL, nil)
		if err != nil {
			log.Printf("Error creating request: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 设置请求头
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Referer", "https://mp.weixin.qq.com/")

		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error fetching content: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// 处理响应内容
		var reader io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				log.Printf("Error creating gzip reader: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer gzReader.Close()
			reader = gzReader
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 设置响应头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'; img-src * data:; default-src 'self' 'unsafe-inline' 'unsafe-eval' https://*.weixin.qq.com https://*.qpic.cn")
		w.Header().Del("Content-Encoding") // 移除 gzip 编码头

		// 处理内容并返回
		content := string(body)
		content = strings.ReplaceAll(content, `data-src="https://mmbiz.qpic.cn/`, `src="/wx-images/`)
		content = strings.ReplaceAll(content, `src="https://mmbiz.qpic.cn/`, `src="/wx-images/`)
		content = strings.ReplaceAll(content, `data-src="https://mmbiz.qlogo.cn/`, `src="/wx-qim/`)
		content = strings.ReplaceAll(content, `src="https://mmbiz.qlogo.cn/`, `src="/wx-qim/`)
		content = strings.ReplaceAll(content, `href="https://mp.weixin.qq.com/`, `href="/wx-mp/`)

		// 添加基础样式和错误处理
		// 添加新窗口打开按钮
		originalURL := targetURL
		openButton := fmt.Sprintf(`
			<div style="position: fixed; top: 20px; right: 20px; z-index: 1000;">
				<a href="%s" target="_blank" style="
					display: inline-block;
					padding: 10px 20px;
					background-color: #07C160;
					color: white;
					border-radius: 5px;
					text-decoration: none;
					font-size: 14px;
					box-shadow: 0 2px 4px rgba(0,0,0,0.1);
					transition: background-color 0.3s;
				">
					新窗口打开
				</a>
			</div>
		`, originalURL)

		content = strings.ReplaceAll(content, `<head>`, `<head>
			<base target="_blank">
			<style>
				img { max-width: 100%; height: auto; }
				img[src=""] { display: none; }
				body { padding: 20px; }
				.rich_media_content { font-size: 16px; line-height: 1.6; }
			</style>
			<script>
				window.onerror = function(msg, url, line) {
					console.error('Error: ' + msg + '\nURL: ' + url + '\nLine: ' + line);
					return false;
				};
			</script>`)
		content = strings.ReplaceAll(content, `<body>`, `<body>`+openButton)

		// 设置响应头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")

		// 返回修改后的内容
		if _, err := w.Write([]byte(content)); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	// 添加微信资源代理路由
	http.HandleFunc("/wx-images/", func(w http.ResponseWriter, r *http.Request) {
		proxyURL := "https://mmbiz.qpic.cn" + strings.TrimPrefix(r.URL.Path, "/wx-images")
		proxyRequest(w, r, proxyURL)
	})

	http.HandleFunc("/wx-qim/", func(w http.ResponseWriter, r *http.Request) {
		proxyURL := "https://mmbiz.qlogo.cn" + strings.TrimPrefix(r.URL.Path, "/wx-qim")
		proxyRequest(w, r, proxyURL)
	})

	http.HandleFunc("/wx-mp/", func(w http.ResponseWriter, r *http.Request) {
		proxyURL := "https://mp.weixin.qq.com" + strings.TrimPrefix(r.URL.Path, "/wx-mp")
		proxyRequest(w, r, proxyURL)
	})

	// 静态文件服务
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// 首页
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/templates/index.html")
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// 通用代理处理函数
func proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	req, err := http.NewRequest(r.Method, targetURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 添加微信相关请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 MicroMessenger/7.0.20.1781(0x6700143B) NetType/WIFI MiniProgramEnv/Windows WindowsWechat/WMPF WindowsWechat(0x6309092b) XWEB/9053")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://mp.weixin.qq.com/")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	// 处理响应内容
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer gzReader.Close()
		reader = gzReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'; img-src * data:; default-src 'self' 'unsafe-inline' 'unsafe-eval' https://*.weixin.qq.com https://*.qpic.cn")
	w.Header().Del("Content-Encoding") // 移除 gzip 编码头

	// 返回处理后的内容
	w.Write(body)
}
