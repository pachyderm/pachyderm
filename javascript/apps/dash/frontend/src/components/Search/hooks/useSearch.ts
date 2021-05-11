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
    reset,
    history,
    setHistory,
  } = useContext(SearchContext);

  const closeDropdown = useCallback(() => {
    setIsOpen(false);
  }, [setIsOpen]);

  const openDropdown = useCallback(() => {
    setIsOpen(true);
  }, [setIsOpen]);

  const addToSearchHistory = useCallback(
    (value: string) => {
      if (
        history.length < MAX_HISTORY &&
        !history.find((item) => item === value)
      ) {
        setHistory((prevHistory) => [...prevHistory, value]);
      }
    },
    [history, setHistory],
  );

  const clearSearchHistory = useCallback(() => {
    setHistory([]);
  }, [setHistory]);

  return {
    isOpen,
    closeDropdown,
    openDropdown,
    searchValue,
    setSearchValue,
    clearSearch: reset,
    history,
    addToSearchHistory,
    clearSearchHistory,
    debouncedValue,
  };
};
