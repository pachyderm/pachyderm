import {PureCheckbox} from '@pachyderm/components';
import classnames from 'classnames';
import {format, fromUnixTime} from 'date-fns';
import React, {
  CSSProperties,
  memo,
  useCallback,
  useEffect,
  useMemo,
  useRef,
} from 'react';
import {areEqual} from 'react-window';

import {GetLogsQuery} from '@graphqlTypes';

import styles from './LogRow.module.css';

type LogRowProps = {
  index: number;
  style: CSSProperties;
  width: number;
  logs: GetLogsQuery['logs'];
  selectedLogsMap: {[key: number]: boolean};
  setSelectedLogsMap: React.Dispatch<
    React.SetStateAction<{[key: number]: boolean}>
  >;
  highlightUserLogs: boolean;
  setSize: (index: number, size: number) => void;
};

const LogRow: React.FC<LogRowProps> = ({
  index,
  style,
  width,
  logs,
  selectedLogsMap,
  setSelectedLogsMap,
  highlightUserLogs,
  setSize,
}) => {
  const heightRef = useRef<null | HTMLDivElement>(null);
  const user = logs[index]?.user;
  const timestamp = logs[index]?.timestamp;
  const message = logs[index]?.message;

  useEffect(() => {
    if (heightRef.current) {
      setSize(index, heightRef.current.getBoundingClientRect().height);
    }
  }, [index, setSize, width]);

  const formattedTimestamp = useMemo(() => {
    if (timestamp) {
      return format(fromUnixTime(timestamp.seconds), 'MM-dd-yyyy H:mm:ss');
    }
    return '';
  }, [timestamp]);

  const onTimestampSelect = useCallback(() => {
    setSelectedLogsMap((prev) => {
      return {...prev, [index]: !prev[index]};
    });
  }, [index, setSelectedLogsMap]);

  return (
    <div
      className={classnames(styles.row, styles[index % 2 ? 'even' : 'odd'])}
      style={style}
    >
      <div className={styles.timestampCol}>
        <PureCheckbox
          className={styles.timestampCheckbox}
          selected={selectedLogsMap[index] || false}
          onChange={onTimestampSelect}
          label={formattedTimestamp}
        />
      </div>
      <div ref={heightRef} className={styles.messageCol}>
        {highlightUserLogs && user ? (
          <mark className={styles.mark}>{message}</mark>
        ) : (
          message
        )}
      </div>
    </div>
  );
};
export default memo(LogRow, areEqual);
