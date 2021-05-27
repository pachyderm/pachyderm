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
