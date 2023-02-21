import {useCallback} from 'react';

import {useSearch} from './useSearch';

export const useDefaultDropdown = () => {
  const {setSearchValue} = useSearch();

  const handleHistoryChipClick = useCallback(
    (value?: string) => {
      if (value) {
        setSearchValue(value);
      }
    },
    [setSearchValue],
  );

  return {handleHistoryChipClick};
};
