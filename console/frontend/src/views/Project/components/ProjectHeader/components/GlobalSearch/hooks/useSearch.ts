import {useCallback, useContext} from 'react';

import {MAX_HISTORY} from '../constants/searchConstants';
import SearchContext from '../contexts/SearchContext';

export const useSearch = () => {
  const {
    isOpen,
    setIsOpen,
    searchValue,
    setSearchValue,
    debouncedValue,
    history,
    setHistory,
  } = useContext(SearchContext);

  const openDropdown = useCallback(() => {
    setIsOpen(true);
  }, [setIsOpen]);

  const addToSearchHistory = useCallback(
    (value: string) => {
      const index = history.findIndex((i) => i === value);
      if (index < 0) {
        if (history.length < MAX_HISTORY) {
          setHistory([value, ...history]);
        } else {
          setHistory([value, ...history].slice(0, -1));
        }
      } else {
        history.splice(index, 1);
        setHistory([value, ...history]);
      }
    },
    [history, setHistory],
  );

  const clearSearchHistory = useCallback(() => {
    setHistory([]);
  }, [setHistory]);

  const clearSearch = useCallback(() => {
    setSearchValue('');
  }, [setSearchValue]);

  return {
    isOpen,
    openDropdown,
    searchValue,
    setSearchValue,
    clearSearch,
    history,
    addToSearchHistory,
    clearSearchHistory,
    debouncedValue,
  };
};
