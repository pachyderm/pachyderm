import {GetLogsQuery} from '@graphqlTypes';
import classnames from 'classnames';
import fill from 'lodash/fill';
import React, {useCallback, useEffect, useState} from 'react';

import {
  DefaultDropdown,
  DropdownItem,
  SkeletonDisplayText,
  PureCheckbox,
} from '@pachyderm/components';

import styles from './LogsListHeader.module.css';

type LogsListHeaderProps = {
  dropdownOptions: DropdownItem[];
  selectedTime: string;
  rawLogs: boolean;
  setSelectedTime: React.Dispatch<React.SetStateAction<string>>;
  selectedLogsMap: {[key: number]: boolean};
  setSelectedLogsMap: React.Dispatch<
    React.SetStateAction<{[key: number]: boolean}>
  >;
  logs: GetLogsQuery['logs'];
  loading: boolean;
};

const LogsListHeader: React.FC<LogsListHeaderProps> = ({
  dropdownOptions,
  selectedTime,
  rawLogs,
  setSelectedTime,
  selectedLogsMap,
  setSelectedLogsMap,
  logs,
  loading,
}) => {
  const [selectAllCheckbox, setSelectAllCheckbox] = useState(false);

  useEffect(() => {
    if (logs.length === 0) {
      setSelectAllCheckbox(false);
    }
    if (Object.values(selectedLogsMap).includes(false)) {
      setSelectAllCheckbox(false);
    }
    if (rawLogs) {
      setSelectAllCheckbox(false);
    }
  }, [logs.length, rawLogs, selectedLogsMap]);

  const onSelectAll = useCallback(() => {
    if (!selectAllCheckbox) {
      setSelectedLogsMap(Object.assign({}, fill(Array(logs.length), true)));
    } else {
      setSelectedLogsMap({});
    }
    setSelectAllCheckbox(!selectAllCheckbox);
  }, [logs.length, selectAllCheckbox, setSelectedLogsMap]);
  return (
    <div
      className={classnames(styles.bodyHeader, {[styles.rawHeader]: rawLogs})}
    >
      <div className={styles.timestampHeader}>
        {!rawLogs && (
          <PureCheckbox
            data-testid="LogsListHeader__select_all"
            selected={selectAllCheckbox}
            disabled={logs.length === 0}
            onChange={onSelectAll}
          />
        )}

        <div
          className={classnames(
            styles[rawLogs ? 'rawTimestampText' : 'timestampText'],
          )}
        >
          {!rawLogs && <span>TIMESTAMP</span>}

          {loading ? (
            <SkeletonDisplayText />
          ) : (
            <DefaultDropdown
              className={styles.datePicker}
              items={dropdownOptions}
              onSelect={setSelectedTime}
            >
              {selectedTime}
            </DefaultDropdown>
          )}
        </div>
      </div>
      {!rawLogs && <div className={styles.messageHeader}>RESPONSE</div>}
    </div>
  );
};

export default LogsListHeader;
