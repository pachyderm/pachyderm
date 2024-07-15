import {useCallback, useState} from 'react';

const useTimestampPagination = (defaultPageSize: number) => {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(defaultPageSize);
  const [cursors, setCursors] = useState<(string | undefined)[]>([undefined]);

  const updateCursor = (cursor?: string) => {
    if (cursor && page + 1 > cursors.length) {
      const cursorIn = cursor;
      setCursors((arr) => [...arr, cursorIn]);
    }
  };

  const resetCursors = useCallback(() => setCursors([undefined]), []);

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
