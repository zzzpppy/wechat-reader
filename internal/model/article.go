package model

import "time"

type Article struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Author      string    `json:"author"`
    Content     string    `json:"content"`
    URL         string    `json:"url"`
    Topic       string    `json:"topic"`      // 添加主题字段
    PublishTime time.Time `json:"publish_time"`
    CreateTime  time.Time `json:"create_time"`
}