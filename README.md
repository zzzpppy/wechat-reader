# 微信阅读器

一个简单的微信公众号文章阅读和管理工具。

## 功能特点

- 支持获取微信公众号文章
- 按主题分类管理文章
- 在线阅读文章内容
- 支持图片资源的正确显示
- 支持新窗口打开原文

## 技术栈

- 后端：Go
- 前端：原生 JavaScript + HTML + CSS
- 数据库：SQLite

## 快速开始

### 环境要求

- Go 1.16+
- SQLite 3

### 安装

1. 克隆项目

```bash
git clone https://github.com/zzzpppy/wechat-reader
cd wechat-reader
```

2. 安装依赖
```bash
go mod tidy
```

3. 运行项目
```bash
go run cmd/server/main.go
```

4. 访问应用
打开浏览器访问 http://localhost:8080

## 使用说明
1. 在输入框中粘贴微信公众号文章链接
2. 点击"获取文章"按钮
3. 文章会自动保存并按主题分类
4. 点击左侧主题可以筛选文章
5. 点击"阅读原文"可以在线阅读文章内容

## 项目结构
```plaintext
wechat-reader/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── service/
│   │   └── crawler.go
│   └── storage/
│       └── database.go
├── web/
│   ├── static/
│   │   └── css/
│   └── templates/
│       └── index.html
└── README.md
 ```

## 注意事项
- 本工具仅用于学习和研究使用
- 请遵守微信公众平台相关规则
- 不要频繁抓取相同的文章链接
## 许可证
MIT License

## 贡献指南
欢迎提交 Issue 和 Pull Request