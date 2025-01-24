import React from 'react';
import { Paper, Typography, Grid, Card, CardContent, CardActions, Button } from '@mui/material';

function ArticleList({ articles = [], onArticleClick }) {
  if (!Array.isArray(articles)) return null;

  return (
    <Paper sx={{ p: 2 }}>
      <Typography variant="h6" gutterBottom>
        文章列表 ({articles.length})
      </Typography>
      <Grid container spacing={2}>
        {articles.map((article) => (
          <Grid item xs={12} key={article.id}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  {article.title || '无标题'}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {article.author && `作者: ${article.author}`}
                  {article.create_time && (
                    <>
                      <br />
                      创建时间: {new Date(article.create_time).toLocaleString()}
                    </>
                  )}
                </Typography>
              </CardContent>
              <CardActions>
                <Button size="small" color="primary" onClick={() => onArticleClick(article.url)}>
                  阅读原文
                </Button>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Paper>
  );
}

export default ArticleList;