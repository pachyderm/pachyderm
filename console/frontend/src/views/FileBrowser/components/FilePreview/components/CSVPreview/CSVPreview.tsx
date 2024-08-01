import React, {CSSProperties, useCallback} from 'react';
import {FixedSizeGrid} from 'react-window';

import {FixedGridRowProps} from '@dash-frontend/lib/types';
import {LoadingDots, DefaultDropdown} from '@pachyderm/components';
import usePanelModal from '@pachyderm/components/Modal/FullPagePanelModal/hooks/usePanelModal';

import styles from './CSVPreview.module.css';
import useCSVPreview, {
  FilePreviewProps,
  DELIMITER_ITEMS,
} from './hooks/useCSVPreview';

const ITEM_HEIGHT = 34;
const ITEM_WIDTH = 160;
const HEADER_OVERHEAD = 170;
const SIDEPANEL_OVERHEAD = 146;
const SIDEPANEL_WIDTH = 252;

const CSVPreview: React.FC<FilePreviewProps> = ({downloadLink}) => {
  const {headers, data, loading, delimiter, setDelimiter, delimiterLabel} =
    useCSVPreview({
      downloadLink,
    });
  const {leftOpen, rightOpen} = usePanelModal();

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

  const containerWidth =
    window.innerWidth -
    SIDEPANEL_OVERHEAD -
    (leftOpen ? SIDEPANEL_WIDTH : 0) -
    (rightOpen ? SIDEPANEL_WIDTH : 0);

  const elementWidth = Math.max(
    ITEM_WIDTH,
    containerWidth / (headers.length || 1) - 6,
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
        width={containerWidth}
        height={window.innerHeight - HEADER_OVERHEAD}
        className={styles.base}
      >
        {Row}
      </FixedSizeGrid>
    </>
  );
};

export default CSVPreview;
