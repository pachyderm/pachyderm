import {ExitSVG, SearchSVG} from '@pachyderm/components';
import classNames from 'classnames';
import React, {useMemo} from 'react';
import {useFormContext} from 'react-hook-form';

import {useSearch} from '../../hooks/useSearch';

import styles from './SearchInput.module.css';

const placeholderText = 'Search for repos, pipelines and jobs';

const SearchInput: React.FC = () => {
  const {isOpen, openDropdown, searchValue, clearSearch} = useSearch();

  const {register} = useFormContext();

  const showButton = useMemo(() => isOpen && searchValue, [
    isOpen,
    searchValue,
  ]);

  return (
    <>
      <span className={styles.searchIcon}>
        <SearchSVG aria-hidden width={22} height={22} />
      </span>
      <input
        name="search"
        role="searchbox"
        placeholder={placeholderText}
        ref={register}
        className={classNames(styles.input, {[styles.open]: isOpen})}
        onFocus={openDropdown}
      />
      {showButton && (
        <button
          className={styles.button}
          aria-label={'Clear search input'}
          type="button"
          onClick={clearSearch}
        >
          <ExitSVG aria-hidden width={13} height={13} />
        </button>
      )}
    </>
  );
};

export default SearchInput;
