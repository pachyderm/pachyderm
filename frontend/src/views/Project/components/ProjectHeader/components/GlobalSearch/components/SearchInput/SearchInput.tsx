import classNames from 'classnames';
import React, {useMemo} from 'react';
import {UseFormReturn} from 'react-hook-form';

import {CloseSVG, SearchSVG, Icon} from '@pachyderm/components';

import {useSearch} from '../../hooks/useSearch';

import styles from './SearchInput.module.css';

const placeholderText = 'Search for repos, pipelines and jobs';

type SearchInputProps = {
  formContext: UseFormReturn;
};

const SearchInput: React.FC<SearchInputProps> = ({formContext}) => {
  const {isOpen, openDropdown, searchValue, clearSearch} = useSearch();

  const showButton = useMemo(
    () => isOpen && searchValue,
    [isOpen, searchValue],
  );

  return (
    <>
      <Icon className={styles.searchIcon} small>
        <SearchSVG aria-hidden />
      </Icon>
      <input
        autoComplete="off"
        role="searchbox"
        placeholder={placeholderText}
        className={classNames(styles.input, {[styles.open]: isOpen})}
        onFocus={openDropdown}
        {...formContext.register('project_search')}
      />
      {showButton && (
        <button
          className={styles.button}
          aria-label={'Clear search input'}
          type="button"
          onClick={clearSearch}
        >
          <Icon color="white" small>
            <CloseSVG aria-hidden />
          </Icon>
        </button>
      )}
    </>
  );
};

export default SearchInput;
