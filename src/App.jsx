import React, { useState, useEffect } from 'react';
import { Container, Typography, TextField, Button, Grid, Paper, CircularProgress, Alert } from '@mui/material';
import TopicList from './components/TopicList';
import ArticleList from './components/ArticleList';
import ArticleModal from './components/ArticleModal';

function App() {
  const [articles, setArticles] = useState([]);
  const [topics, setTopics] = useState(['全部']);
  const [selectedTopic, setSelectedTopic] = useState('全部');
  const [urlInput, setUrlInput] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [currentArticleUrl, setCurrentArticleUrl] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadTopics();
    loadArticles();
  }, []);

  const loadTopics = async () => {
    try {
      const response = await fetch('/api/topics');
      const data = await response.json();
      setTopics(['全部', ...(data.data || [])]);
    } catch (error) {
      console.error('加载主题失败:', error);
      setError('加载主题失败，请刷新页面重试');
    }
  };

  const loadArticles = async () => {
    try {
      const response = await fetch('/api/articles');
      const data = await response.json();
      setArticles(data.data || []);
    } catch (error) {
      console.error('加载文章失败:', error);
      setError('加载文章失败，请刷新页面重试');
    } finally {
      setIsLoading(false);
    }
  };

  const fetchArticles = async () => {
    if (!urlInput.trim()) {
      alert('请输入文章链接');
      return;
    }

    setIsLoading(true);
    try {
      const response = await fetch('/api/fetch', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'Accept-Language': 'zh-CN,zh;q=0.9'
        },
        body: JSON.stringify({ url: urlInput }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      if (result.success && Array.isArray(result.data)) {
        setUrlInput('');
        setArticles(prevArticles => [...prevArticles, ...result.data]);
        await loadTopics();
        setSelectedTopic('全部');
      } else {
        throw new Error(result.message || '获取文章失败');
      }
    } catch (error) {
      console.error('获取文章时发生错误:', error);
      alert(error.message || '获取文章失败，请检查网络连接');
    } finally {
      setIsLoading(false);
    }
  };

  const handleOpenArticle = (url) => {
    setCurrentArticleUrl(url);
    setModalOpen(true);
  };

  const filteredArticles = selectedTopic === '全部'
    ? articles
    : articles.filter(article => article.topic === selectedTopic);

  if (isLoading) {
    return (
      <Container sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress />
      </Container>
    );
  }

  if (error) {
    return (
      <Container>
        <Alert severity="error" sx={{ mt: 2 }}>
          {error}
        </Alert>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Typography variant="h3" component="h1" gutterBottom>
        微信阅读器
      </Typography>

      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2} alignItems="center">
          <Grid item xs>
            <TextField
              fullWidth
              placeholder="输入微信文章链接"
              value={urlInput}
              onChange={(e) => setUrlInput(e.target.value)}
            />
          </Grid>
          <Grid item>
            <Button 
              variant="contained" 
              onClick={fetchArticles}
              disabled={isLoading}
            >
              {isLoading ? '获取中...' : '获取文章'}
            </Button>
          </Grid>
        </Grid>
      </Paper>

      <Grid container spacing={3}>
        <Grid item xs={12} md={3}>
          <TopicList
            topics={topics}
            selectedTopic={selectedTopic}
            onTopicSelect={setSelectedTopic}
            articles={articles}
          />
        </Grid>
        <Grid item xs={12} md={9}>
          <ArticleList
            articles={filteredArticles}
            onArticleClick={handleOpenArticle}
          />
        </Grid>
      </Grid>

      <ArticleModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        articleUrl={currentArticleUrl}
      />
    </Container>
  );
}

export default App;