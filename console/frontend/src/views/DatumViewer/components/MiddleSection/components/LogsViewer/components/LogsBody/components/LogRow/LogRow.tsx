import classnames from 'classnames';
import React, {
  CSSProperties,
  memo,
  useCallback,
  useEffect,
  useMemo,
  useRef,
} from 'react';
import {areEqual} from 'react-window';

import {LogMessage} from '@dash-frontend/api/pps';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {CodeText, PureCheckbox} from '@pachyderm/components';

import styles from './LogRow.module.css';

type LogRowProps = {
  index: number;
  style: CSSProperties;
  width: number;
  logs?: LogMessage[];
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
  const user = logs?.[index]?.user;
  const timestamp = logs?.[index]?.ts;
  const message = logs?.[index]?.message;

  useEffect(() => {
    if (heightRef.current) {
      setSize(index, heightRef.current.getBoundingClientRect().height);
    }
  }, [index, setSize, width]);

  const formattedTimestamp = useMemo(() => {
    if (timestamp) {
      return getStandardDateFromISOString(timestamp);
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
      data-testid="LogRow__base"
      className={classnames(styles.row, styles[index % 2 ? 'even' : 'odd'])}
      style={style}
    >
      <div className={styles.timestampCol}>
        <PureCheckbox
          data-testid="LogRow__checkbox"
          className={styles.timestampCheckbox}
          selected={selectedLogsMap[index] || false}
          onChange={onTimestampSelect}
        />
      </div>
      <div ref={heightRef} className={styles.messageCol}>
        {highlightUserLogs && user ? (
          <CodeText data-testid="LogRow__user_log" className={styles.mark}>
            {formattedTimestamp} {message}
          </CodeText>
        ) : (
          <CodeText>
            {formattedTimestamp} {message}
          </CodeText>
        )}
      </div>
    </div>
  );
};
export default memo(LogRow, areEqual);
