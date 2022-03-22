import {LoadingDots} from '@pachyderm/components';
import intersection from 'lodash/intersection';
import React, {CSSProperties} from 'react';
import {FixedSizeList} from 'react-window';

import {FixedListRowProps} from '@dash-frontend/lib/types';
import {
  ITEM_SIZE,
  PREVIEW_WIDTH,
  HEADER_OVERHEAD,
  PADDING_SIZE,
} from '@dash-frontend/views/FileBrowser/constants/FileBrowser';

import ContentWrapper from '../ContentWrapper';

import JSONRow from './components/JSONRow';
import useJSONPreview from './hooks/useJSONPreview';

const InnerElement = ({style, ...rest}: {style: CSSProperties}) => (
  <ul
    style={{
      ...style,
      height: `${parseFloat(style.height + '') + PADDING_SIZE * 2}px`,
    }}
    {...rest}
  />
);

type FilePreviewProps = {
  downloadLink: string;
};

const JSONPreview: React.FC<FilePreviewProps> = ({downloadLink}) => {
  const {flatData, loading, minimizedRows, setMinimizedRows} =
    useJSONPreview(downloadLink);

  const Row: React.FC<FixedListRowProps> = ({index, style}) => {
    const {childOf, id} = flatData[index];

    const contentHidden = intersection(minimizedRows, childOf).length > 0;

    if (contentHidden) return null;

    return (
      <JSONRow
        itemData={flatData[index]}
        style={style}
        handleMinimize={() =>
          setMinimizedRows((prevMinimizedRows) => [...prevMinimizedRows, id])
        }
        handleExpand={() =>
          setMinimizedRows((prevMinimizedRows) =>
            prevMinimizedRows.filter((v) => v !== id),
          )
        }
      />
    );
  };

  if (loading)
    return (
      <span data-testid="JSONPreview__loading">
        <LoadingDots />
      </span>
    );

  return (
    <ContentWrapper>
      <FixedSizeList
        height={window.innerHeight - HEADER_OVERHEAD}
        itemCount={flatData.length}
        innerElementType={InnerElement}
        itemSize={ITEM_SIZE}
        width={PREVIEW_WIDTH}
      >
        {Row}
      </FixedSizeList>
    </ContentWrapper>
  );
};

export default JSONPreview;
