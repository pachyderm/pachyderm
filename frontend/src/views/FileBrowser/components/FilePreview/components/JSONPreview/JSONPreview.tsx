import React from 'react';

import JSONDataPreview from '@dash-frontend/components/JSONDataPreview';
import {LoadingDots} from '@pachyderm/components';

import ContentWrapper from '../ContentWrapper';

import useJSONPreview from './hooks/useJSONPreview';

type FilePreviewProps = {
  downloadLink: string;
};

const JSONPreview: React.FC<FilePreviewProps> = ({downloadLink}) => {
  const {data, loading} = useJSONPreview(downloadLink);

  if (loading)
    return (
      <span data-testid="JSONPreview__loading">
        <LoadingDots />
      </span>
    );

  return (
    <ContentWrapper>
      <JSONDataPreview inputData={data} formattingStyle="json" />
    </ContentWrapper>
  );
};

export default JSONPreview;
