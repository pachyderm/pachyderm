import React from 'react';

import ContentWrapper from '../ContentWrapper';

import styles from './IFramePreview.module.css';

type FilePreviewProps = {
  downloadLink: string;
  fileName: string;
  'data-testid'?: string;
};

const IFramePreview: React.FC<FilePreviewProps> = ({
  downloadLink,
  fileName,
  'data-testid': dataTestId,
}) => {
  return (
    <ContentWrapper>
      <iframe
        data-testid={dataTestId || 'IFramePreview__iframe'}
        className={styles.base}
        src={downloadLink}
        title={fileName}
        sandbox=""
      />
    </ContentWrapper>
  );
};

export default IFramePreview;
