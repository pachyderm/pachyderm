import {
  stringComparator,
  useSort,
  numberComparator,
} from '@pachyderm/components';
import {useCallback} from 'react';

import {File} from '@graphqlTypes';

const nameComparator = {
  name: 'Name',
  func: stringComparator,
  accessor: (record: File) => record.path,
};

const sizeComparator = {
  name: 'Size',
  func: numberComparator,
  accessor: (record: File) => record.sizeBytes,
};

const dateComparator = {
  name: 'Date',
  func: numberComparator,
  accessor: (record: File) => record.committed?.seconds || 0,
};

const typeComparator = {
  name: 'Type',
  func: stringComparator,
  accessor: (record: File) =>
    record.path.slice(record.path.lastIndexOf('.') + 1),
};

const useListViewTable = (files: File[]) => {
  const {sortedData, setComparator, reversed, comparatorName} = useSort({
    data: files,
    initialSort: nameComparator,
    initialDirection: 1,
  });

  const nameClick = useCallback(() => {
    setComparator(nameComparator);
  }, [setComparator]);

  const sizeClick = useCallback(() => {
    setComparator(sizeComparator);
  }, [setComparator]);

  const dateClick = useCallback(() => {
    setComparator(dateComparator);
  }, [setComparator]);

  const typeClick = useCallback(() => {
    setComparator(typeComparator);
  }, [setComparator]);

  return {
    comparatorName,
    nameClick,
    sizeClick,
    dateClick,
    typeClick,
    reversed,
    tableData: sortedData,
  };
};

export default useListViewTable;
