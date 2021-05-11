import {Form, useOutsideClick} from '@pachyderm/components';
import classnames from 'classnames';
import React, {useCallback, useMemo, useRef, useState} from 'react';
import {useForm} from 'react-hook-form';

import DefaultDropdown from './components/DefaultDropdown';
import SearchBar from './components/SearchInput';
import SearchResults from './components/SearchResultsDropdown';
import SearchContext from './contexts/SearchContext';
import {useDebounce} from './hooks/useDebounce';
import {useSearchHistory} from './hooks/useSearchHistory';
import styles from './Search.module.css';

const Search: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const {history, setHistory} = useSearchHistory();

  const formCtx = useForm();
  const {watch, setValue, reset} = formCtx;
  const debouncedValue = useDebounce(watch('search'), 200);
  const searchValue = watch('search');

  const setSearchValue = useCallback(
    (value) => {
      setValue('search', value);
    },
    [setValue],
  );

  const ctxValue = useMemo(
    () => ({
      isOpen,
      setIsOpen,
      searchValue,
      setSearchValue,
      debouncedValue,
      reset,
      history,
      setHistory,
    }),
    [
      isOpen,
      setIsOpen,
      searchValue,
      setSearchValue,
      debouncedValue,
      reset,
      history,
      setHistory,
    ],
  );

  const showResults = searchValue !== undefined && searchValue !== '';
  const showDefaultDropdown = !searchValue;

  const containerRef = useRef<HTMLDivElement>(null);
  const handleOutsideClick = useCallback(() => {
    if (isOpen) {
      setIsOpen(false);
    }
  }, [isOpen, setIsOpen]);
  useOutsideClick(containerRef, handleOutsideClick);

  return (
    <SearchContext.Provider value={ctxValue}>
      <div className={styles.base} ref={containerRef}>
        <Form formContext={formCtx}>
          <SearchBar />
        </Form>
        <div
          className={classnames(styles.dropdown, {
            [styles.open]: isOpen,
          })}
        >
          {showDefaultDropdown && <DefaultDropdown />}
          {showResults && <SearchResults />}
        </div>
      </div>
    </SearchContext.Provider>
  );
};

export default Search;
