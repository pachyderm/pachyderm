import {LoadingDots} from '@pachyderm/components';
import React from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';
import {VariableSizeList} from 'react-window';

import EmptyState from '@dash-frontend/components/EmptyState';
import {GetLogsQuery} from '@graphqlTypes';

import {HEADER_HEIGHT_OFFSET} from '../../../constants/logsViewersConstants';
import useLogsBody from '../../hooks/useLogsBody';

import LogRow from './components';

type LogsBodyProps = {
  logs: GetLogsQuery['logs'];
  loading: boolean;
  highlightUserLogs: boolean;
  selectedLogsMap: {[key: string]: boolean};
  setSelectedLogsMap: React.Dispatch<
    React.SetStateAction<{[key: string]: boolean}>
  >;
};

const LogsBody: React.FC<LogsBodyProps> = ({
  logs,
  loading,
  highlightUserLogs,
  selectedLogsMap,
  setSelectedLogsMap,
}) => {
  const {listRef, getSize, setSize} = useLogsBody();

  if (loading) return <LoadingDots />;
  if (logs.length > 0) {
    return (
      <AutoSizer>
        {({height, width}) => (
          <VariableSizeList
            ref={listRef}
            width={width}
            height={height - HEADER_HEIGHT_OFFSET}
            itemCount={logs.length}
            itemSize={getSize}
            overscanCount={4}
          >
            {({...props}) => (
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
          </VariableSizeList>
        )}
      </AutoSizer>
    );
  }

  return (
    <>
      <EmptyState
        title="No logs found for this time range."
        message="Try adjusting the filter above."
      />
    </>
  );
};

export default LogsBody;
