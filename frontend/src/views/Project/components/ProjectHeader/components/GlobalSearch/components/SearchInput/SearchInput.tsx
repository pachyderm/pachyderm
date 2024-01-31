import classNames from 'classnames';
import React, {useMemo} from 'react';
import {UseFormReturn} from 'react-hook-form';

import {CloseSVG, SearchSVG, Icon} from '@pachyderm/components';

import {useSearch} from '../../hooks/useSearch';

import styles from './SearchInput.module.css';

const placeholderText = 'Search Project';

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
    <div
      className={classNames(styles.inputWrapper, {
        [styles.backgroundGrey9]: isOpen,
      })}
    >
      <Icon className={styles.searchIcon} small color="white">
        <SearchSVG aria-hidden />
      </Icon>
      <div className={styles.underline}>
        <input
          autoComplete="off"
          role="searchbox"
          placeholder={placeholderText}
          className={classNames(styles.input, {
            [styles.open]: isOpen,
            [styles.backgroundGrey9]: isOpen,
          })}
          onFocus={openDropdown}
          {...formContext.register('project_search')}
        />
      </div>
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
    </div>
  );
};

export default SearchInput;
