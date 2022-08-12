import {LoadingDots, DefaultDropdown} from '@pachyderm/components';
import React, {CSSProperties, useCallback} from 'react';
import {FixedSizeGrid} from 'react-window';

import {FixedGridRowProps} from '@dash-frontend/lib/types';

import styles from './CSVPreview.module.css';
import useCSVPreview, {
  FilePreviewProps,
  DELIMITER_ITEMS,
} from './hooks/useCSVPreview';

const ITEM_HEIGHT = 34;
const ITEM_WIDTH = 160;
const HEADER_OVERHEAD = 170;

const CSVPreview: React.FC<FilePreviewProps> = ({downloadLink}) => {
  const {headers, data, loading, delimiter, setDelimiter, delimiterLabel} =
    useCSVPreview({
      downloadLink,
    });

  const Row: React.FC<FixedGridRowProps> = ({rowIndex, columnIndex, style}) => {
    const rowValue = data[rowIndex];
    const columnValue = headers[columnIndex];
    return (
      <div
        className={styles.cell}
        style={{
          ...style,
          top: `${parseFloat(style.top + '') + ITEM_HEIGHT}px`,
        }}
      >
        {JSON.stringify(rowValue[columnValue])}
      </div>
    );
  };

  const elementWidth = Math.max(
    ITEM_WIDTH,
    window.innerWidth / (headers.length || 1) - 5,
  );

  const InnerElement = useCallback(
    ({style, ...rest}: {style: CSSProperties}) => (
      <div className={styles.content}>
        <header className={styles.header}>
          {headers.map((key) => (
            <strong
              className={styles.cell}
              key={key}
              style={{
                width: `${elementWidth}px`,
              }}
            >
              {key}
            </strong>
          ))}
        </header>
        <div
          style={{
            ...style,
            height: `${parseFloat(style.height + '') + ITEM_HEIGHT}px`,
          }}
          {...rest}
        />
      </div>
    ),
    [elementWidth, headers],
  );

  if (loading)
    return (
      <span data-testid="CSVPreview__loading">
        <LoadingDots />
      </span>
    );

  return (
    <>
      <div className={styles.settingsHeader}>
        <DefaultDropdown
          initialSelectId={delimiter}
          onSelect={setDelimiter}
          menuOpts={{pin: 'right'}}
          items={DELIMITER_ITEMS}
        >
          Separator: {delimiterLabel}
        </DefaultDropdown>
      </div>
      <FixedSizeGrid
        columnWidth={elementWidth}
        columnCount={headers.length}
        rowCount={data.length}
        rowHeight={ITEM_HEIGHT}
        innerElementType={InnerElement}
        width={window.innerWidth}
        height={window.innerHeight - HEADER_OVERHEAD}
        className={styles.base}
      >
        {Row}
      </FixedSizeGrid>
    </>
  );
};

export default CSVPreview;
