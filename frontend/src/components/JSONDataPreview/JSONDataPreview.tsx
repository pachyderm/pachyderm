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

import FixedRow from './components/FixedRow';
import useDataPreview from './hooks/useDataPreview';

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
  inputData: unknown;
  formattingStyle?: 'json' | 'yaml';
  width?: number;
  height?: number;
};

const JSONDataPreview: React.FC<FilePreviewProps> = ({
  inputData,
  formattingStyle = 'json',
  width,
  height,
}) => {
  const {flatData, minimizedRows, setMinimizedRows} = useDataPreview(
    inputData,
    formattingStyle === 'json',
  );

  const Row: React.FC<FixedListRowProps> = ({index, style}) => {
    const {childOf, id} = flatData[index];

    const contentHidden = intersection(minimizedRows, childOf).length > 0;

    if (contentHidden) return null;

    return (
      <FixedRow
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

  return (
    <FixedSizeList
      height={height || window.innerHeight - HEADER_OVERHEAD}
      itemCount={flatData.length}
      innerElementType={InnerElement}
      itemSize={ITEM_SIZE}
      width={width || PREVIEW_WIDTH}
    >
      {Row}
    </FixedSizeList>
  );
};

export default JSONDataPreview;
