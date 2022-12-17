import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import {getDatumStateColor, getDatumStateSVG} from '@dash-frontend/lib/datums';
import {
  CloseSVG,
  Icon,
  LoadingDots,
  Search,
  StatusStopSVG,
} from '@pachyderm/components';

import ListItem from '../ListItem';

import styles from './DatumList.module.css';
import useDatumList from './hooks/useDatumList';

export type DatumListProps = {
  setIsExpanded: React.Dispatch<React.SetStateAction<boolean>>;
};

const DatumList: React.FC<DatumListProps> = ({setIsExpanded}) => {
  const {
    datums,
    currentDatumId,
    onDatumClick,
    loading,
    searchValue,
    setSearchValue,
    clearSearch,
    showNoSearchResults,
  } = useDatumList(setIsExpanded);

  if (loading) {
    return <LoadingDots />;
  }

  if (!loading && datums.length === 0) {
    return <EmptyState title="No datums found for this job." />;
  }

  return (
    <div className={styles.base} data-testid="DatumList__list">
      <div className={styles.header}>
        <div className={styles.search}>
          <Search
            data-testid="DatumList__search"
            value={searchValue}
            placeholder="Enter the exact datum ID"
            onSearch={setSearchValue}
            className={styles.searchInput}
          />
          {searchValue && (
            <Icon small color="black" className={styles.searchClose}>
              <CloseSVG
                onClick={clearSearch}
                data-testid="DatumList__searchClear"
              />
            </Icon>
          )}
        </div>

        {showNoSearchResults && (
          <ListItem
            LeftIconSVG={StatusStopSVG}
            text="No matching datums found"
          />
        )}
      </div>

      {datums &&
        datums.map((datum) => (
          <ListItem
            data-testid="DatumList__listItem"
            key={datum.id}
            state={currentDatumId === datum.id ? 'selected' : 'default'}
            LeftIconSVG={getDatumStateSVG(datum.state) || undefined}
            leftIconColor={getDatumStateColor(datum.state) || undefined}
            text={datum.id}
            onClick={() => onDatumClick(datum.id)}
          />
        ))}
    </div>
  );
};
export default DatumList;
