package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
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
	// 验证URL是否为微信文章链接
	if !strings.Contains(subscriptionURL, "mp.weixin.qq.com") {
		return nil, fmt.Errorf("无效的微信文章链接")
	}

	// 创建带有适当请求头的请求
	req, err := http.NewRequest("GET", subscriptionURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置必要的请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 MicroMessenger/7.0.20.1781(0x6700143B) NetType/WIFI MiniProgramEnv/Windows WindowsWechat/WMPF WindowsWechat(0x6309092b) XWEB/9053")
	// req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept", "text/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://mp.weixin.qq.com/")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// 打印请求信息
	fmt.Printf("发送请求到: %s\n", subscriptionURL)
	fmt.Printf("请求头: %+v\n", req.Header)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 打印响应状态
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应头: %+v\n", resp.Header)

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	// 处理 gzip 压缩
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("解压 gzip 内容失败: %v", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// 读取响应内容
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应内容是否为空
	if len(body) == 0 {
		return nil, fmt.Errorf("服务器返回空响应")
	}

	fmt.Println(string(body))

	// 解析HTML内容
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	// 获取主题信息
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

	msgid := ""
	itemidx := 0

	// 从 DOM 中提取文章列表
	var articles []model.Article
	if len(articles) == 0 {
		// 从文章列表中提取文章信息
		doc.Find(".album__list.js_album_list .album__list-item.js_album_item.js_wx_tap_highlight.wx_tap_cell").Each(func(i int, s *goquery.Selection) {
			id := fmt.Sprintf("article_%d_%d", time.Now().Unix(), i)

			// 从 data-link 和 data-title 属性获取文章信息
			link, exists := s.Attr("data-link")
			if !exists || !strings.Contains(link, "mp.weixin.qq.com") {
				return
			}

			title, exists := s.Attr("data-title")
			if !exists || title == "" {
				title = fmt.Sprintf("未命名文章_%d", i+1)
			}
			title = strings.TrimSpace(title)

			// 获取 msgid、itemidx（如果有的话）
			msgidT, exexists := s.Attr("data-msgid")
			if !exexists {
				msgid = ""
			}
			msgid = msgidT
			itemidxT, exexists := s.Attr("data-itemidx")
			if !exexists {
				itemidx = 0
			}
			itemidx, _ = strconv.Atoi(itemidxT)

			// 获取发布时间（如果有的话）
			publishTime := time.Now() // 默认使用当前时间
			publishTimeStr := strings.TrimSpace(s.Find(".album__item-info-time").Text())
			if publishTimeStr != "" {
				if t, err := time.Parse("2006-01-02", publishTimeStr); err == nil {
					publishTime = t
				}
			}

			article := model.Article{
				ID:          id,
				Title:       title,
				URL:         link,
				Topic:       topic,
				PublishTime: publishTime,
				CreateTime:  time.Now(),
			}
			articles = append(articles, article)
			fmt.Printf("从文章列表解析到文章: %+v\n", article)
		})

		// 如果从文章列表中没有找到文章，尝试从其他链接中查找
		if len(articles) == 0 {
			doc.Find("a[href*='mp.weixin.qq.com']").Each(func(i int, s *goquery.Selection) {
				id := fmt.Sprintf("article_%d_%d", time.Now().Unix(), i)
				link, exists := s.Attr("href")
				if !exists || strings.Contains(link, "javascript:") {
					return
				}

				title := strings.TrimSpace(s.Text())
				if title == "" {
					title = fmt.Sprintf("未命名文章_%d", i+1)
				}

				article := model.Article{
					ID:          id,
					Title:       title,
					URL:         link,
					Topic:       topic,
					PublishTime: time.Now(),
					CreateTime:  time.Now(),
				}
				articles = append(articles, article)
				fmt.Printf("从其他链接解析到文章: %+v\n", article)
			})
		}
	}

	// 在获取完初始文章后，尝试获取更多文章
	if len(articles) > 0 {
		// 从 URL 中提取 topic_id
		topicID := ""
		if matches := regexp.MustCompile(`album_id=([^&]+)`).FindStringSubmatch(subscriptionURL); len(matches) > 1 {
			topicID = matches[1]

			// 获取更多文章
			moreArticles, err := c.fetchMoreArticles(ctx, topicID, topic, msgid, itemidx)
			if err != nil {
				fmt.Printf("获取更多文章时出错: %v\n", err)
			} else {
				articles = append(articles, moreArticles...)
			}
		}
	}

	// 打印最终结果
	fmt.Printf("总共解析到 %d 篇文章\n", len(articles))
	for i, article := range articles {
		fmt.Printf("文章 %d: %+v\n", i+1, article)
	}
	return articles, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 在 Crawler struct 定义后添加以下内容
type WeixinResponse struct {
	BaseResp struct {
		Ret int `json:"ret"`
	} `json:"base_resp"`
	GetalbumResp struct {
		ArticleList  []WeixinArticle `json:"article_list"`
		ContinueFlag string          `json:"continue_flag"`
	} `json:"getalbum_resp"`
}

type WeixinArticle struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	CoverImg   string `json:"cover_img_1_1"`
	CreateTime string `json:"create_time"`
	Msgid      string `json:"msgid"`
	Itemidx    string `json:"itemidx"`
}

func (c *Crawler) fetchMoreArticles(ctx context.Context, topicID string, topic string, msgid string, itemidex int) ([]model.Article, error) {
	var allArticles []model.Article
	processedURLs := make(map[string]bool)

	nextMsgid := ""
	nextItemidx := 0
	if msgid != "" {
		nextMsgid = msgid
	}
	if itemidex != 0 {
		nextItemidx = itemidex
	}
	hasMore := true
	batchSize := 10

	for hasMore {
		url := "https://mp.weixin.qq.com/mp/appmsgalbum"
		params := map[string]string{
			"action":   "getalbum",
			"album_id": topicID,
			"count":    fmt.Sprintf("%d", batchSize),
			"f":        "json",
		}

		if nextMsgid != "" && nextItemidx > 0 {
			params["begin_msgid"] = nextMsgid
			params["begin_itemidx"] = fmt.Sprintf("%d", nextItemidx)
		}

		// 构建请求 URL
		reqURL := url + "?"
		for k, v := range params {
			reqURL += k + "=" + v + "&"
		}
		reqURL = strings.TrimSuffix(reqURL, "&")

		// 创建请求
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Referer", fmt.Sprintf("https://mp.weixin.qq.com/mp/appmsgalbum?action=getalbum&album_id=%s", topicID))

		// 发送请求
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		// 解析响应
		var result WeixinResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}

		if result.BaseResp.Ret != 0 {
			return nil, fmt.Errorf("API返回错误码: %d", result.BaseResp.Ret)
		}

		newArticlesCount := 0
		for _, wxArticle := range result.GetalbumResp.ArticleList {
			if !processedURLs[wxArticle.URL] {
				// 将字符串类型的创建时间转换为int64
				createTimeInt, err := strconv.ParseInt(wxArticle.CreateTime, 10, 64)
				if err != nil {
					// 如果转换失败，使用当前时间作为发布时间
					createTimeInt = time.Now().Unix()
				}

				article := model.Article{
					ID:          fmt.Sprintf("article_%d_%d", time.Now().Unix(), len(allArticles)),
					Title:       wxArticle.Title,
					URL:         wxArticle.URL,
					Topic:       topic,
					PublishTime: time.Unix(createTimeInt, 0),
					CreateTime:  time.Now(),
				}
				allArticles = append(allArticles, article)
				processedURLs[wxArticle.URL] = true
				newArticlesCount++
			}
		}

		if newArticlesCount == 0 || len(result.GetalbumResp.ArticleList) == 0 {
			break
		}

		// 更新下一次请求的参数
		lastArticle := result.GetalbumResp.ArticleList[len(result.GetalbumResp.ArticleList)-1]
		nextMsgid = lastArticle.Msgid
		itemidx, err := strconv.Atoi(lastArticle.Itemidx)
		if err != nil {
			nextItemidx = 0
		} else {
			nextItemidx = itemidx
		}

		cf, err := strconv.ParseInt(result.GetalbumResp.ContinueFlag, 10, 64)
		if err != nil {
		}
		if cf == 0 {
			hasMore = false
		}

		// 添加延时避免被封
		time.Sleep(2 * time.Second)
	}

	return allArticles, nil
}
