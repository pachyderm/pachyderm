import {
  DefaultDropdown,
  DropdownItem,
  SkeletonDisplayText,
} from '@pachyderm/components';
import fill from 'lodash/fill';
import React, {useCallback, useEffect, useState} from 'react';

import {GetLogsQuery} from '@graphqlTypes';

import styles from './LogsListHeader.module.css';

type LogsListHeaderProps = {
  dropdownOptions: DropdownItem[];
  selectedTime: string;
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
  }, [logs.length, selectedLogsMap]);

  const onSelectAll = useCallback(() => {
    if (!selectAllCheckbox) {
      setSelectedLogsMap(Object.assign({}, fill(Array(logs.length), true)));
    } else {
      setSelectedLogsMap({});
    }
    setSelectAllCheckbox(!selectAllCheckbox);
  }, [logs.length, selectAllCheckbox, setSelectedLogsMap]);

  return (
    <div className={styles.bodyHeader}>
      <div className={styles.timestampHeader}>
        <input
          type="checkbox"
          checked={selectAllCheckbox}
          disabled={logs.length === 0}
          onChange={onSelectAll}
        />
        <div className={styles.timestampText}>
          <span>TIMESTAMP</span>

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
      <div className={styles.messageHeader}>RESPONSE</div>
    </div>
  );
};

export default LogsListHeader;
