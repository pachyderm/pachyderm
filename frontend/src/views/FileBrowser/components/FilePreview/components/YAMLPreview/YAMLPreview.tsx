import React from 'react';

import JSONDataPreview from '@dash-frontend/components/JSONDataPreview';
import {LoadingDots} from '@pachyderm/components';

import ContentWrapper from '../ContentWrapper';

import useYAMLPreview from './hooks/useYAMLPreview';

type FilePreviewProps = {
  downloadLink: string;
};

const YAMLPreview: React.FC<FilePreviewProps> = ({downloadLink}) => {
  const {data, loading} = useYAMLPreview(downloadLink);

  if (loading)
    return (
      <span data-testid="YAMLPreview__loading">
        <LoadingDots />
      </span>
    );

  return (
    <ContentWrapper>
      <JSONDataPreview inputData={data} formattingStyle="yaml" />
    </ContentWrapper>
  );
};

export default YAMLPreview;
