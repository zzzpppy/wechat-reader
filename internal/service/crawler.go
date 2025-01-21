package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"wechat-reader/internal/model"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	client *http.Client
}

func NewCrawler() *Crawler {
	return &Crawler{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *Crawler) FetchArticles(ctx context.Context, subscriptionURL string) ([]model.Article, error) {
	// 添加调试用的请求头
	req, err := http.NewRequest("GET", subscriptionURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 打印请求信息
	fmt.Printf("正在请求 URL: %s\n", subscriptionURL)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 打印响应状态
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应头: %v\n", resp.Header)

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 打印响应内容的一部分
	// fmt.Printf("响应内容前500字符: %s\n", string(body[:min(len(body), 500)]))
	fmt.Printf("body: %s\n", string(body))

	// 重新创建 Reader 用于 goquery
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	// 打印整个 HTML 结构中的主题相关信息
	doc.Find("#js_tag_name").Each(func(i int, s *goquery.Selection) {
		html, err := s.Html()
		if err != nil {
			fmt.Printf("获取主题容器HTML失败: %v\n", err)
			return
		}
		fmt.Printf("找到主题容器: %s\n", html)
	})

    var topic string
    doc.Find("#js_tag_name").Each(func(i int, s *goquery.Selection) {
        // 获取完整的文本内容
        fullText := strings.TrimSpace(s.Text())
        // 移除可能的前缀和多余空格
        topic = strings.TrimSpace(strings.TrimPrefix(fullText, ""))
        fmt.Printf("提取到的主题文本: %s\n", topic)
    })

    if topic == "" {
        topic = "未分类" // 设置默认主题
    }

    var articles []model.Article
    // 移除文章数量限制，获取所有文章链接
    doc.Find("a[data-link]").Each(func(i int, s *goquery.Selection) {
        id := fmt.Sprintf("article_%d_%d", time.Now().Unix(), i)
        dataLink, exists := s.Attr("data-link")
        if !exists {
            return
        }

        article := model.Article{
            ID:          id,
            Title:       s.Text(),
            URL:         dataLink,
            Topic:       topic,
            PublishTime: time.Now(),
            CreateTime:  time.Now(),
        }

        articles = append(articles, article)
    })

	return articles, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
