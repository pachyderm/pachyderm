import React from 'react';

import {LoadingDots} from '@pachyderm/components';

import MarkdownDataPreview from './components/MarkdownDataPreview';
import useMarkdownPreview from './hooks/useMarkdownPreview';
import styles from './MarkdownPreview.module.css';

type MarkdownPreviewProps = {
  downloadLink: string;
};

const MarkdownPreview: React.FC<MarkdownPreviewProps> = ({downloadLink}) => {
  const {data, loading} = useMarkdownPreview(downloadLink);

  if (loading) <LoadingDots />;

  return (
    <div className={styles.markdownPreview}>
      <MarkdownDataPreview markdown={data || ''} />
    </div>
  );
};

export default MarkdownPreview;
