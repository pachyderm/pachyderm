import React from 'react';

import {Button, Chip, ChipGroup} from '@pachyderm/components';

import {useDefaultDropdown} from '../../hooks/useDefaultDropdown';
import {useSearch} from '../../hooks/useSearch';
import {NotFoundMessage, SectionHeader} from '../Messaging';

import styles from './DefaultDropdown.module.css';

const DefaultDropdown: React.FC = () => {
  const {history, clearSearchHistory} = useSearch();
  const {handleHistoryChipClick} = useDefaultDropdown();

  const recentSearch = () => {
    if (history.length > 0) {
      return (
        <>
          <div className={styles.sectionHeader}>
            <SectionHeader>Recent Searches</SectionHeader>
            <Button onClick={clearSearchHistory} buttonType="ghost">
              Clear
            </Button>
          </div>
          <div className={styles.recentSearchGroup}>
            <ChipGroup>
              {history.map((searchValue) => (
                <Chip
                  key={searchValue}
                  onClickValue={searchValue}
                  onClick={handleHistoryChipClick}
                >
                  {searchValue}
                </Chip>
              ))}
            </ChipGroup>
          </div>
        </>
      );
    }

    return <NotFoundMessage>There are no recent searches.</NotFoundMessage>;
  };

  return <div className={styles.base}>{recentSearch()}</div>;
};

export default DefaultDropdown;
