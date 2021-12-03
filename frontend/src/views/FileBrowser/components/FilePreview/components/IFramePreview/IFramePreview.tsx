import React from 'react';

import ContentWrapper from '../ContentWrapper';

import styles from './IFramePreview.module.css';

type FilePreviewProps = {
  downloadLink: string;
  fileName: string;
};

const IFramePreview: React.FC<FilePreviewProps> = ({
  downloadLink,
  fileName,
}) => {
  return (
    <ContentWrapper>
      <iframe
        data-testid="IFramePreview__iframe"
        className={styles.base}
        src={downloadLink}
        title={fileName}
        sandbox=""
      />
    </ContentWrapper>
  );
};

export default IFramePreview;
