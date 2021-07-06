import {useCallback, useRef} from 'react';
import {VariableSizeList} from 'react-window';

import {DEFAULT_ROW_HEIGHT} from '../../constants/logsViewersConstants';

const useLogsBody = () => {
  const listRef = useRef<VariableSizeList>(null);
  const sizeMap = useRef<{[key: string]: number}>({});

  const setSize = useCallback((index: number, size: number) => {
    if (sizeMap.current[index] !== size) {
      sizeMap.current = {...sizeMap.current, [index]: size};
      if (listRef.current) {
        listRef.current.resetAfterIndex(0);
      }
    }
  }, []);

  const getSize = useCallback((index) => {
    return sizeMap.current[index] || DEFAULT_ROW_HEIGHT;
  }, []);

  return {
    listRef,
    setSize,
    getSize,
  };
};
export default useLogsBody;
