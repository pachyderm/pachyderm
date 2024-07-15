import React from 'react';
import ReactMarkdown from 'react-markdown';
import gfm from 'remark-gfm';

import styles from './MarkdownDataPreview.module.css';

type MarkdownDataPreviewProps = {
  markdown: string;
};

const MarkdownDataPreview: React.FC<MarkdownDataPreviewProps> = ({
  markdown,
}) => {
  return (
    <div className={styles.markdownWrapper}>
      <ReactMarkdown remarkPlugins={[gfm]}>{markdown}</ReactMarkdown>
    </div>
  );
};

export default MarkdownDataPreview;
