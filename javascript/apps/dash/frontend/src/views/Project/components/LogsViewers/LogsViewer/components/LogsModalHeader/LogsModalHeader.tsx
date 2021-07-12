import {
  Button,
  ButtonLink,
  CopySVG,
  Switch,
  useClipboardCopy,
} from '@pachyderm/components';
import {format, fromUnixTime} from 'date-fns';
import React, {useCallback, useEffect, useState} from 'react';

import Header from '@dash-frontend/components/Header';
import {GetLogsQuery} from '@graphqlTypes';

import {LOGS_DATE_FORMAT} from '../../../constants/logsViewersConstants';
import useDownloadText from '../../hooks/useDownloadText';

import styles from './LogsModalHeader.module.css';

type LogsModalHeaderProps = {
  setHighlightUserLogs: React.Dispatch<React.SetStateAction<boolean>>;
  selectedLogsMap: {[key: number]: boolean};
  logs: GetLogsQuery['logs'];
  headerText: string;
};

const LogsModalHeader: React.FC<LogsModalHeaderProps> = ({
  setHighlightUserLogs,
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
      <h4 className={styles.title}>{headerText}</h4>
      <div className={styles.controls}>
        <div className={styles.switchGroup}>
          <Switch className={styles.switch} onChange={setHighlightUserLogs} />
          <span>Highlight User Logs</span>
        </div>
        <div className={styles.exportGroup}>
          <ButtonLink
            className={styles.copy}
            onClick={copy}
            disabled={disableExport}
            small
          >
            <CopySVG /> Copy selected rows
          </ButtonLink>
          <Button onClick={download} disabled={disableExport}>
            Download
          </Button>
        </div>
      </div>
    </Header>
  );
};

export default LogsModalHeader;
