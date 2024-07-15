import React, {useCallback} from 'react';

import {LoadingDots} from '@pachyderm/components';
import {useFetch} from 'hooks/useFetch';

import ContentWrapper from '../ContentWrapper';

import styles from './WebPreview.module.css';

type FilePreviewProps = {
  downloadLink: string;
  fileName: string;
};

const WebPreview: React.FC<FilePreviewProps> = ({downloadLink, fileName}) => {
  const formatResponse = useCallback(
    async (res: Response) => await res.text(),
    [],
  );
  const {data, loading} = useFetch<string>({
    url: downloadLink,
    formatResponse,
  });

  if (loading) return <LoadingDots />;

  return (
    <ContentWrapper>
      <iframe
        data-testid="WebPreview__iframe"
        className={styles.base}
        srcDoc={JSON.stringify(data)}
        title={fileName}
        sandbox=""
      />
    </ContentWrapper>
  );
};

export default WebPreview;
