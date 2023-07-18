import {ApolloError} from '@apollo/client';
import {GetLogsQuery} from '@graphqlTypes';
import classnames from 'classnames';
import React from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';

import BrandedDocLink from '@dash-frontend/components/BrandedDocLink';
import EmptyState from '@dash-frontend/components/EmptyState';
import {LoadingDots} from '@pachyderm/components';

import {HEADER_HEIGHT_OFFSET} from '../../constants/logsViewersConstants';
import useLogsBody from '../../hooks/useLogsBody';

import LogRow from './components/LogRow';
import LogsList from './components/LogsList';
import RawLogRow from './components/RawLogRow';
import styles from './LogsBody.module.css';

type LogsBodyProps = {
  logs: GetLogsQuery['logs']['items'];
  loading: boolean;
  highlightUserLogs: boolean;
  rawLogs: boolean;
  selectedLogsMap: {[key: string]: boolean};
  setSelectedLogsMap: React.Dispatch<
    React.SetStateAction<{[key: string]: boolean}>
  >;
  error?: ApolloError;
  isSkippedDatum?: boolean;
};

const LogsBody: React.FC<LogsBodyProps> = ({
  logs,
  loading,
  highlightUserLogs,
  rawLogs,
  selectedLogsMap,
  setSelectedLogsMap,
  error,
  isSkippedDatum,
}) => {
  const {listRef, getSize, setSize, isDatum} = useLogsBody();

  if (loading) return <LoadingDots />;
  if (logs.length > 0) {
    return (
      <AutoSizer>
        {({height, width}) => (
          <LogsList
            className={classnames({[styles.raw]: rawLogs})}
            forwardRef={listRef}
            width={width}
            height={height - HEADER_HEIGHT_OFFSET}
            itemCount={logs.length}
            itemSize={getSize}
            overscanCount={4}
          >
            {({...props}) => (
              <>
                {rawLogs ? (
                  <RawLogRow
                    {...props}
                    width={width}
                    logs={logs}
                    highlightUserLogs={highlightUserLogs}
                    setSize={setSize}
                  />
                ) : (
                  <LogRow
                    {...props}
                    width={width}
                    logs={logs}
                    selectedLogsMap={selectedLogsMap}
                    setSelectedLogsMap={setSelectedLogsMap}
                    highlightUserLogs={highlightUserLogs}
                    setSize={setSize}
                  />
                )}
              </>
            )}
          </LogsList>
        )}
      </AutoSizer>
    );
  }

  if (error) {
    return (
      <EmptyState
        title={`No logs found for this ${isDatum ? 'datum' : 'job'}.`}
      >
        If you haven&apos;t already, consider setting up a persistent log
        aggregator like{' '}
        <BrandedDocLink pathWithoutDomain="deploy-manage/deploy/loki/">
          Loki
        </BrandedDocLink>
      </EmptyState>
    );
  }

  if (isSkippedDatum) {
    return (
      <EmptyState
        title="Skipped datum."
        message="This datum has been successfully processed in a previous job."
      />
    );
  }
  return (
    <EmptyState
      title="No logs found for this time range."
      message="Try adjusting the filter above."
    />
  );
};

export default LogsBody;
