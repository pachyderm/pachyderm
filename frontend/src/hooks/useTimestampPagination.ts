import {TimestampInput, Timestamp} from '@graphqlTypes';
import {useCallback, useState} from 'react';

const useTimestampPagination = (defaultPageSize: number) => {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(defaultPageSize);
  const [cursors, setCursors] = useState<(TimestampInput | null)[]>([null]);

  const updateCursor = (cursor?: Timestamp | null) => {
    if (cursor && page + 1 > cursors.length) {
      const cursorIn = {seconds: cursor.seconds, nanos: cursor.nanos};
      setCursors((arr) => [...arr, cursorIn]);
    }
  };

  const resetCursors = useCallback(() => setCursors([null]), []);

  return {
    page,
    setPage,
    pageSize,
    setPageSize,
    cursors,
    updateCursor,
    resetCursors,
  };
};

export default useTimestampPagination;
