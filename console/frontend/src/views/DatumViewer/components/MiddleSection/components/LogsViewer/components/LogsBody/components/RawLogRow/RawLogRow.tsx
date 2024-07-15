import classnames from 'classnames';
import React, {CSSProperties, memo, useEffect, useRef} from 'react';
import {areEqual} from 'react-window';

import {LogMessage} from '@dash-frontend/api/pps';
import {CodeText} from '@pachyderm/components';

import styles from './RawLogRow.module.css';

type LogRowProps = {
  index: number;
  style: CSSProperties;
  width: number;
  logs?: LogMessage[];
  highlightUserLogs: boolean;
  setSize: (index: number, size: number) => void;
};

const RawLogRow: React.FC<LogRowProps> = ({
  index,
  style,
  width,
  logs,
  highlightUserLogs,
  setSize,
}) => {
  const heightRef = useRef<null | HTMLDivElement>(null);
  const user = logs?.[index]?.user;
  const message = logs?.[index]?.message;

  useEffect(() => {
    if (heightRef.current) {
      setSize(
        index,
        heightRef.current.getBoundingClientRect().height +
          (index === 0 || (logs?.length && index === logs.length - 1) ? 16 : 0),
      );
    }
  }, [index, logs?.length, setSize, width]);

  return (
    <div
      data-testid="RawLogRow__base"
      className={classnames(styles.row, {[styles.firstRow]: index === 0})}
      style={style}
    >
      <div ref={heightRef} className={styles.messageCol}>
        {highlightUserLogs && user ? (
          <CodeText data-testid="RawLogRow__user_log" className={styles.mark}>
            {message}
          </CodeText>
        ) : (
          <CodeText className={styles.rawText}>{message}</CodeText>
        )}
      </div>
    </div>
  );
};
export default memo(RawLogRow, areEqual);
