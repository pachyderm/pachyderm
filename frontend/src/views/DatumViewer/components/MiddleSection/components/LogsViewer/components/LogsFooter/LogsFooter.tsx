import {QueryObserverResult, RefetchOptions} from '@tanstack/react-query';
import React, {useState} from 'react';

import {LogMessage} from '@dash-frontend/api/pps';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {ActionUpdateSVG, Button, SimplePager} from '@pachyderm/components';

import {LOGS_PAGE_SIZE} from '../../constants/logsViewersConstants';

import styles from './LogsFooter.module.css';

type LogsFooterProps = {
  logs?: LogMessage[];
  refetch: (
    options?: RefetchOptions | undefined,
  ) => Promise<QueryObserverResult<LogMessage[], Error>>;
  page: number;
  setPage: React.Dispatch<React.SetStateAction<number>>;
};

const LogsFooter: React.FC<LogsFooterProps> = ({
  logs,
  refetch,
  page,
  setPage,
}) => {
  const hasNextPage = logs?.length && logs.length >= LOGS_PAGE_SIZE;

  const footerText = logs?.[0]?.ts
    ? getStandardDateFromISOString(logs?.[0]?.ts)
    : null;

  const [loading, setLoading] = useState(false);

  const refresh = async () => {
    setLoading(true);
    await refetch();

    const timer = setTimeout(() => {
      setLoading(false);
    }, 1000);

    return () => clearTimeout(timer);
  };

  return (
    <div className={styles.base}>
      <div className={styles.left}>
        {footerText && `Viewing logs from ${footerText}`}
      </div>

      <div className={styles.right}>
        <div className={styles.divider}>
          <Button
            onClick={refresh}
            disabled={hasNextPage || loading}
            iconPosition="end"
            IconSVG={ActionUpdateSVG}
            buttonType="ghost"
            color="black"
          >
            Refresh
          </Button>
        </div>
        <SimplePager
          page={page}
          updatePage={setPage}
          nextPageDisabled={!hasNextPage}
          pageSize={LOGS_PAGE_SIZE}
          contentLength={0}
        />
      </div>
    </div>
  );
};

export default LogsFooter;
