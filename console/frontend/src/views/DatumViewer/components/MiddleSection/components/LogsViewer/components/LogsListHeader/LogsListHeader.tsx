import fill from 'lodash/fill';
import React, {useCallback, useEffect, useState} from 'react';

import {LogMessage} from '@dash-frontend/api/pps';
import {PureCheckbox} from '@pachyderm/components';

import styles from './LogsListHeader.module.css';

type LogsListHeaderProps = {
  rawLogs: boolean;
  selectedLogsMap: {[key: number]: boolean};
  setSelectedLogsMap: React.Dispatch<
    React.SetStateAction<{[key: number]: boolean}>
  >;
  logs?: LogMessage[];
  loading: boolean;
};

const LogsListHeader: React.FC<LogsListHeaderProps> = ({
  rawLogs,
  selectedLogsMap,
  setSelectedLogsMap,
  logs,
}) => {
  const [selectAllCheckbox, setSelectAllCheckbox] = useState(false);

  useEffect(() => {
    if (!logs?.length) {
      setSelectAllCheckbox(false);
    }
    if (Object.values(selectedLogsMap).includes(false)) {
      setSelectAllCheckbox(false);
    }
    if (rawLogs) {
      setSelectAllCheckbox(false);
    }
  }, [logs?.length, rawLogs, selectedLogsMap]);

  const onSelectAll = useCallback(() => {
    if (!selectAllCheckbox) {
      setSelectedLogsMap(Object.assign({}, fill(Array(logs?.length), true)));
    } else {
      setSelectedLogsMap({});
    }
    setSelectAllCheckbox(!selectAllCheckbox);
  }, [logs?.length, selectAllCheckbox, setSelectedLogsMap]);
  return (
    <div className={styles.bodyHeader}>
      {!rawLogs && (
        <>
          <div className={styles.timestampHeader}>
            {!rawLogs && (
              <PureCheckbox
                className={styles.checkbox}
                data-testid="LogsListHeader__select_all"
                selected={selectAllCheckbox}
                disabled={!logs?.length}
                onChange={onSelectAll}
              />
            )}
            <span className={styles.timestampText}>Timestamp</span>
          </div>
          Message
        </>
      )}
    </div>
  );
};

export default LogsListHeader;
