import {GetLogsQuery} from '@graphqlTypes';
import classnames from 'classnames';
import React, {useEffect, useMemo, useState} from 'react';
import {UseFormReturn} from 'react-hook-form';

import useDownloadText from '@dash-frontend/hooks/useDownloadText';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {
  Button,
  ButtonGroup,
  CaptionTextSmall,
  Checkbox,
  CopySVG,
  DownloadSVG,
  Dropdown,
  FilterSVG,
  Form,
  RadioButton,
  Tooltip,
  useClipboardCopy,
} from '@pachyderm/components';

import {LOGS_DEFAULT_DROPDOWN_OPTIONS} from '../../constants/logsViewersConstants';
import {LogsViewerFormValues} from '../../hooks/useLogsViewer';

import styles from './LogsControls.module.css';

type LogsControlsProps = {
  selectedLogsMap: {[key: number]: boolean};
  logs: GetLogsQuery['logs'];

  formCtx: UseFormReturn<LogsViewerFormValues>;
};

const LogsControls: React.FC<LogsControlsProps> = ({
  selectedLogsMap,
  logs,
  formCtx,
}) => {
  const {pipelineId, datumId} = useUrlState();

  const [disableExport, setDisableExport] = useState(true);

  useEffect(() => {
    setDisableExport(
      logs.length === 0 || !Object.values(selectedLogsMap).includes(true),
    );
  }, [logs.length, selectedLogsMap]);

  const formatedText = useMemo(() => {
    return Object.entries(selectedLogsMap)
      .reduce((acc: string[], [index, selected]) => {
        if (selected) {
          const message = logs[Number(index)]?.message;
          const timestamp = logs[Number(index)]?.timestamp;
          acc.push(
            `${timestamp ? getStandardDate(timestamp.seconds) : '-'} ${
              message || ''
            }`,
          );
        }
        return acc;
      }, [])
      .join('\n');
  }, [logs, selectedLogsMap]);

  const {copy} = useClipboardCopy(formatedText);
  const {download} = useDownloadText(
    formatedText,
    datumId ? `${pipelineId}_datum_logs` : `${pipelineId}_logs`,
  );
  return (
    <ButtonGroup>
      <Tooltip
        tooltipKey="download"
        tooltipText="Download selected"
        placement="bottom"
        disabled={disableExport}
      >
        <Button
          IconSVG={DownloadSVG}
          onClick={download}
          disabled={disableExport}
          buttonType="ghost"
          color="black"
          aria-label="Download selected logs"
        />
      </Tooltip>

      <Tooltip
        tooltipKey="copy"
        tooltipText="Copy selected"
        placement="bottom"
        disabled={disableExport}
      >
        <Button
          disabled={disableExport}
          IconSVG={CopySVG}
          onClick={copy}
          buttonType="ghost"
          color="black"
          aria-label="Copy selected logs"
        />
      </Tooltip>

      <Dropdown>
        <Tooltip
          tooltipKey="filter"
          tooltipText="Filter logs"
          placement="bottom"
        >
          <Dropdown.Button
            className={styles.dropdownButton}
            IconSVG={FilterSVG}
            color="black"
            buttonType="ghost"
            aria-label="Filter logs"
          />
        </Tooltip>
        <Dropdown.Menu pin="right" className={styles.dropdownMenu}>
          <Form formContext={formCtx}>
            <div className={classnames(styles.divider, styles.group)}>
              <div className={styles.heading}>
                <CaptionTextSmall>View Logs From</CaptionTextSmall>
              </div>

              <RadioButton id="default" name="selectedTime" value="default">
                <RadioButton.Label>
                  {datumId ? 'Datum Start Time' : 'Job Start Time'}
                </RadioButton.Label>
              </RadioButton>
              {LOGS_DEFAULT_DROPDOWN_OPTIONS.map((option) => (
                <RadioButton
                  key={option.id}
                  id={option.id}
                  name="selectedTime"
                  value={option.id}
                >
                  <RadioButton.Label>{option.content}</RadioButton.Label>
                </RadioButton>
              ))}
            </div>
            <div className={styles.group}>
              <div className={styles.heading}>
                <CaptionTextSmall>Display Options</CaptionTextSmall>
              </div>

              <Checkbox
                small
                label="Highlight User Logs"
                name="highlightUserLogs"
              />
              <Checkbox small label="Raw Logs" name="displayRawLogs" />
            </div>
          </Form>
        </Dropdown.Menu>
      </Dropdown>
    </ButtonGroup>
  );
};

export default LogsControls;
