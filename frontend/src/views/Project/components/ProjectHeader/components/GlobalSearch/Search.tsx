import classnames from 'classnames';
import React, {
  useCallback,
  useMemo,
  useRef,
  useState,
  SetStateAction,
} from 'react';
import {useForm} from 'react-hook-form';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Form, useOutsideClick} from '@pachyderm/components';

import DefaultDropdown from './components/DefaultDropdown';
import SearchBar from './components/SearchInput';
import SearchResults from './components/SearchResultsDropdown';
import SearchContext from './contexts/SearchContext';
import {useDebounce} from './hooks/useDebounce';
import styles from './Search.module.css';

const Search: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const {projectId} = useUrlState();
  const [history = [], setHistory] = useLocalProjectSettings({
    projectId,
    key: 'search_history',
  });

  const formCtx = useForm();
  const {watch, setValue} = formCtx;
  const searchValue = watch('project_search');
  const debouncedValue = useDebounce(searchValue, 200);

  const setSearchValue = useCallback(
    (value: SetStateAction<string>) => {
      setValue('project_search', value);
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
      history,
      setHistory,
    }),
    [
      isOpen,
      setIsOpen,
      searchValue,
      setSearchValue,
      debouncedValue,
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
          <SearchBar formContext={formCtx} />
        </Form>
        <div
          data-testid="Search__dropdown"
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
