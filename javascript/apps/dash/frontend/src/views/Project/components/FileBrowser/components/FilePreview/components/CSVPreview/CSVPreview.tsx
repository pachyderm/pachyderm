import {LoadingDots} from '@pachyderm/components';
import {csv} from 'd3-fetch';
import React, {
  CSSProperties,
  useCallback,
  useEffect,
  useState,
  useMemo,
} from 'react';
import {FixedSizeGrid} from 'react-window';

import styles from './CSVPreview.module.css';

const ITEM_HEIGHT = 34;
const ITEM_WIDTH = 160;
const HEADER_OVERHEAD = 170;

type FilePreviewProps = {
  downloadLink: string | undefined | null;
};

const CSVPreview: React.FC<FilePreviewProps> = ({downloadLink}) => {
  const [data, setData] = useState<Record<string, unknown>[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const getData = async () => {
      if (downloadLink) {
        setLoading(true);
        const responseData = await csv(downloadLink, {
          credentials: 'include',
          mode: 'cors',
        });
        responseData && setData(responseData);
        setLoading(false);
      }
    };
    getData();
  }, [downloadLink]);

  const headers = useMemo(() => Object.keys(data[0] || {}), [data]);

  const Row = useCallback(
    ({
      rowIndex,
      columnIndex,
      style,
    }: {
      columnIndex: number;
      rowIndex: number;
      style: CSSProperties;
    }) => {
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
    },
    [data, headers],
  );

  const InnerElement = useCallback(
    ({style, ...rest}: {style: CSSProperties}) => (
      <div className={styles.content}>
        <header className={styles.header}>
          {headers.map((key) => (
            <div className={styles.cell} key={key}>
              {key}
            </div>
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
    [headers],
  );

  if (loading) return <LoadingDots />;

  return (
    <FixedSizeGrid
      columnWidth={ITEM_WIDTH}
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
  );
};

export default CSVPreview;
