import React from 'react';
import { Dialog, DialogContent, IconButton } from '@mui/material';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';

function ArticleModal({ open, onClose, articleUrl }) {
  const getProxiedUrl = (url) => {
    if (!url) return '';
    // 所有微信文章链接都通过代理访问
    return `/api/proxy?url=${encodeURIComponent(url)}`;
  };

  const handleOpenInNewTab = () => {
    if (articleUrl) {
      window.open(articleUrl, '_blank');
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="lg"
      fullWidth
      PaperProps={{
        sx: { height: '90vh', position: 'relative' }
      }}
    >
      <IconButton
        onClick={handleOpenInNewTab}
        sx={{
          position: 'absolute',
          right: 16,
          top: 16,
          zIndex: 1,
          backgroundColor: '#07C160',
          color: 'white',
          '&:hover': {
            backgroundColor: '#06AE56'
          }
        }}
      >
        <OpenInNewIcon />
      </IconButton>
      <DialogContent sx={{ p: 0, height: '100%' }}>
        {articleUrl && (
          <iframe
            src={getProxiedUrl(articleUrl)}
            style={{
              width: '100%',
              height: '100%',
              border: 'none'
            }}
            title="文章内容"
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

export default ArticleModal;