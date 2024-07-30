import React from 'react';

import {Chip} from '@dash-frontend/components/Chip/Chip';
import EmptyState from '@dash-frontend/components/EmptyState';
import ListItem from '@dash-frontend/components/ListItem';
import {getDatumStateColor, getDatumStateSVG} from '@dash-frontend/lib/datums';
import {DATUM_LIST_PAGE_SIZE} from '@dash-frontend/views/DatumViewer/constants/DatumViewer';
import {
  ButtonLink,
  CloseSVG,
  Icon,
  LoadingDots,
  Search,
  SpinnerSVG,
  StatusStopSVG,
  SimplePager,
} from '@pachyderm/components';

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
    page,
    setPage,
    hasNextPage,
    pageCount,
    isProcessing,
    contentLength,
    refresh,
    isSearchValid,
  } = useDatumList(setIsExpanded);

  return (
    <div className={styles.base} data-testid="DatumList__list">
      <div className={styles.header}>
        <SimplePager
          page={page}
          updatePage={setPage}
          nextPageDisabled={!hasNextPage}
          pageCount={pageCount}
          pageSize={DATUM_LIST_PAGE_SIZE}
          contentLength={contentLength}
          elementName="Datum"
        />

        {isProcessing && (
          <Chip
            LeftIconSVG={SpinnerSVG}
            LeftIconSmall={false}
            isButton={false}
            className={styles.loadingMessage}
          >
            <div data-testid="DatumList__processing">
              Processing — datums are being processed.{' '}
              <ButtonLink onClick={() => refresh()}>Refresh</ButtonLink>
            </div>
          </Chip>
        )}

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

        {searchValue && !isSearchValid && (
          <ListItem
            LeftIconSVG={StatusStopSVG}
            text="Enter the exact datum ID"
          />
        )}

        {showNoSearchResults && (
          <ListItem
            LeftIconSVG={StatusStopSVG}
            text="No matching datums found"
          />
        )}
      </div>
      {loading ? (
        <LoadingDots />
      ) : datums?.length !== 0 ? (
        datums?.map((datum) => (
          <ListItem
            data-testid="DatumList__listItem"
            key={datum?.datum?.id}
            state={currentDatumId === datum?.datum?.id ? 'selected' : 'default'}
            LeftIconSVG={getDatumStateSVG(datum.state) || undefined}
            leftIconColor={getDatumStateColor(datum.state) || undefined}
            text={datum?.datum?.id}
            onClick={() => onDatumClick(datum?.datum?.id || '')}
          />
        ))
      ) : (
        <EmptyState title="" message="No datums found for this job." />
      )}
    </div>
  );
};
export default DatumList;
