import {LoadingDots} from '@pachyderm/components';
import intersection from 'lodash/intersection';
import React, {CSSProperties, useCallback} from 'react';
import {FixedSizeList} from 'react-window';

import JSONRow from './components/JSONRow';
import useJSONPreview from './hooks/useJSONPreview';
import styles from './JSONPreview.module.css';

const ITEM_SIZE = 24;
const PREVIEW_WIDTH = 1000;
const HEADER_OVERHEAD = 200;
export const PADDING_SIZE = 32;

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

  const Row = useCallback(
    ({index, style}: {index: number; style: CSSProperties}) => {
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
    },
    [flatData, minimizedRows, setMinimizedRows],
  );

  if (loading) return <LoadingDots />;

  return (
    <FixedSizeList
      height={window.innerHeight - HEADER_OVERHEAD}
      itemCount={flatData.length}
      innerElementType={InnerElement}
      itemSize={ITEM_SIZE}
      width={PREVIEW_WIDTH}
      className={styles.base}
    >
      {Row}
    </FixedSizeList>
  );
};

export default JSONPreview;
