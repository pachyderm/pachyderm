import React from 'react';

import {Button, CaptionTextSmall, Chip, ChipGroup} from '@pachyderm/components';

import {useDefaultDropdown} from '../../hooks/useDefaultDropdown';
import {useSearch} from '../../hooks/useSearch';

import styles from './RecentSearches.module.css';

const RecentSearches: React.FC = () => {
  const {history, clearSearchHistory} = useSearch();
  const {handleHistoryChipClick} = useDefaultDropdown();

  return (
    <>
      {history.length > 0 && (
        <div className={styles.base}>
          <div className={styles.sectionHeader}>
            <CaptionTextSmall>Recent Searches</CaptionTextSmall>
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
        </div>
      )}
    </>
  );
};

export default RecentSearches;
