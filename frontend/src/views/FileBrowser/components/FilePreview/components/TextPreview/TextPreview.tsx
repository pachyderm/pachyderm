import React, {CSSProperties} from 'react';
import {FixedSizeList} from 'react-window';

import {FixedListRowProps} from '@dash-frontend/lib/types';
import {
  ITEM_SIZE,
  PREVIEW_WIDTH,
  HEADER_OVERHEAD,
} from '@dash-frontend/views/FileBrowser/constants/FileBrowser';
import {LoadingDots} from '@pachyderm/components';

import ContentWrapper from '../ContentWrapper';

import TextRow from './components/TextRow';
import useTextPreview from './hooks/useTextPreview';

const InnerElement = ({style, ...rest}: {style: CSSProperties}) => (
  <ol
    style={{
      ...style,
      marginBottom: '0px',
    }}
    {...rest}
  />
);

type FilePreviewProps = {
  downloadLink: string;
};

const TextPreview: React.FC<FilePreviewProps> = ({downloadLink}) => {
  const {data, loading} = useTextPreview(downloadLink);

  const Row: React.FC<FixedListRowProps> = ({index, style}) => {
    return <TextRow text={data[index]} style={style} index={index} />;
  };

  if (loading) return <LoadingDots />;

  return (
    <ContentWrapper>
      <FixedSizeList
        height={window.innerHeight - HEADER_OVERHEAD}
        itemCount={data.length}
        innerElementType={InnerElement}
        itemSize={ITEM_SIZE}
        width={PREVIEW_WIDTH}
      >
        {Row}
      </FixedSizeList>
    </ContentWrapper>
  );
};

export default TextPreview;
