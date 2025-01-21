package storage

import (
	"context"
	"database/sql"
	"time"
	"wechat-reader/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(ctx context.Context, dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 修改这里，不要每次都删除表
	_, err = db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS articles (
            id TEXT PRIMARY KEY,
            title TEXT NOT NULL,
            author TEXT,
            content TEXT,
            url TEXT,
            topic TEXT,
            publish_time DATETIME,
            create_time DATETIME
        )
    `)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close(ctx context.Context) error {
	return d.db.Close()
}

func (d *Database) SaveArticles(ctx context.Context, articles []model.Article) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
        INSERT OR REPLACE INTO articles (id, title, author, content, url, topic, publish_time, create_time)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, article := range articles {
		_, err = stmt.ExecContext(ctx,
			article.ID,
			article.Title,
			article.Author,
			article.Content,
			article.URL,
			article.Topic,
			article.PublishTime,
			article.CreateTime,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *Database) GetArticles(ctx context.Context) ([]model.Article, error) {
	rows, err := d.db.QueryContext(ctx, `
        SELECT id, title, COALESCE(author, ''), COALESCE(content, ''), 
               COALESCE(url, ''), COALESCE(topic, '未分类'), 
               strftime('%Y-%m-%d %H:%M:%S', COALESCE(publish_time, CURRENT_TIMESTAMP)), 
               strftime('%Y-%m-%d %H:%M:%S', COALESCE(create_time, CURRENT_TIMESTAMP))
        FROM articles
        ORDER BY create_time DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		var article model.Article
		var publishTimeStr, createTimeStr string

		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Author,
			&article.Content,
			&article.URL,
			&article.Topic,
			&publishTimeStr,
			&createTimeStr,
		)
		if err != nil {
			return nil, err
		}

		// 解析时间字符串
		article.PublishTime, _ = time.Parse("2006-01-02 15:04:05", publishTimeStr)
		article.CreateTime, _ = time.Parse("2006-01-02 15:04:05", createTimeStr)

		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// 添加获取主题列表的方法
func (d *Database) GetTopics(ctx context.Context) ([]string, error) {
	rows, err := d.db.QueryContext(ctx, `
        SELECT DISTINCT topic 
        FROM articles 
        WHERE topic IS NOT NULL AND topic != ''
        ORDER BY topic
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}

	return topics, nil
}
