import React from 'react';
import { Paper, List, ListItem, ListItemText, Typography } from '@mui/material';

function TopicList({ topics = [], selectedTopic, onTopicSelect, articles = [] }) {
  const getTopicCount = (topic) => {
    if (!Array.isArray(articles)) return 0;
    if (topic === '全部') {
      return articles.length;
    }
    return articles.filter(article => article.topic === topic).length;
  };

  if (!Array.isArray(topics)) return null;

  return (
    <Paper sx={{ p: 2 }}>
      <Typography variant="h6" gutterBottom>
        主题列表
      </Typography>
      <List>
        {topics.map((topic) => (
          <ListItem
            key={topic}
            button
            selected={selectedTopic === topic}
            onClick={() => onTopicSelect(topic)}
            sx={{
              '&.Mui-selected': {
                backgroundColor: 'primary.light',
                '&:hover': {
                  backgroundColor: 'primary.light',
                },
              },
            }}
          >
            <ListItemText
              primary={
                <Typography variant="body1">
                  {topic} ({getTopicCount(topic)})
                </Typography>
              }
            />
          </ListItem>
        ))}
      </List>
    </Paper>
  );
}

export default TopicList;