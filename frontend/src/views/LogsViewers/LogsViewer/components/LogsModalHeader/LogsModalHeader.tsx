import {GetLogsQuery} from '@graphqlTypes';
import {format, fromUnixTime} from 'date-fns';
import React, {useCallback, useEffect, useState} from 'react';

import Header from '@dash-frontend/components/Header';
import {
  Button,
  ButtonGroup,
  CopySVG,
  Switch,
  useClipboardCopy,
} from '@pachyderm/components';

import {LOGS_DATE_FORMAT} from '../../../constants/logsViewersConstants';
import useDownloadText from '../../hooks/useDownloadText';

import styles from './LogsModalHeader.module.css';

type LogsModalHeaderProps = {
  setHighlightUserLogs: React.Dispatch<React.SetStateAction<boolean>>;
  rawLogs: boolean;
  setRawLogs: (selected: boolean) => void;
  selectedLogsMap: {[key: number]: boolean};
  logs: GetLogsQuery['logs'];
  headerText: string;
};

const LogsModalHeader: React.FC<LogsModalHeaderProps> = ({
  setHighlightUserLogs,
  rawLogs,
  setRawLogs,
  selectedLogsMap,
  logs,
  headerText,
}) => {
  const [disableExport, setDisableExport] = useState(true);

  useEffect(() => {
    setDisableExport(
      logs.length === 0 || !Object.values(selectedLogsMap).includes(true),
    );
  }, [logs.length, selectedLogsMap]);

  const formatText = useCallback(() => {
    return Object.entries(selectedLogsMap)
      .reduce((acc: string[], [index, selected]) => {
        if (selected) {
          const message = logs[Number(index)]?.message;
          const timestamp = logs[Number(index)]?.timestamp;
          acc.push(
            `${
              timestamp
                ? format(fromUnixTime(timestamp.seconds), LOGS_DATE_FORMAT)
                : LOGS_DATE_FORMAT.replace(/(.*?)/, ' ')
            } ${message || ''}`,
          );
        }
        return acc;
      }, [])
      .join('\n');
  }, [logs, selectedLogsMap]);

  const {copy} = useClipboardCopy(formatText());
  const {download} = useDownloadText(formatText(), `${headerText}_logs`);
  return (
    <Header appearance="light">
      <h5 className={styles.title}>{headerText}</h5>
      <div className={styles.controls}>
        <div className={styles.switchGroup}>
          <div className={styles.switchItem}>
            <Switch
              className={styles.switch}
              onChange={setHighlightUserLogs}
              aria-label="Highlight User Logs"
            />
            Highlight User Logs
          </div>
          <div className={styles.switchItem}>
            <Switch
              defaultChecked={rawLogs}
              className={styles.switch}
              onChange={setRawLogs}
              aria-label="Raw Logs"
            />
            Raw Logs
          </div>
        </div>
        <ButtonGroup className={styles.exportGroup}>
          <Button
            buttonType="ghost"
            IconSVG={CopySVG}
            className={styles.copy}
            onClick={copy}
            disabled={disableExport}
          >
            Copy selected rows
          </Button>
          <Button onClick={download} disabled={disableExport}>
            Download
          </Button>
        </ButtonGroup>
      </div>
    </Header>
  );
};

export default LogsModalHeader;
