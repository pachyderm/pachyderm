import {CloseSVG, SearchSVG, Icon} from '@pachyderm/components';
import classNames from 'classnames';
import React, {useMemo} from 'react';
import {useFormContext} from 'react-hook-form';

import {useSearch} from '../../hooks/useSearch';

import styles from './SearchInput.module.css';

const placeholderText = 'Search for repos, pipelines and jobs';

const SearchInput: React.FC = () => {
  const {isOpen, openDropdown, searchValue, clearSearch} = useSearch();

  const {register} = useFormContext();

  const showButton = useMemo(
    () => isOpen && searchValue,
    [isOpen, searchValue],
  );

  return (
    <>
      <Icon className={styles.searchIcon}>
        <SearchSVG aria-hidden />
      </Icon>
      <input
        role="searchbox"
        placeholder={placeholderText}
        className={classNames(styles.input, {[styles.open]: isOpen})}
        onFocus={openDropdown}
        {...register('search')}
      />
      {showButton && (
        <button
          className={styles.button}
          aria-label={'Clear search input'}
          type="button"
          onClick={clearSearch}
        >
          <CloseSVG aria-hidden />
        </button>
      )}
    </>
  );
};

export default SearchInput;
