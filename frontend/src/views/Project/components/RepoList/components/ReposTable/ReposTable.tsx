import React from 'react';

import {TableViewFilters} from '@dash-frontend/components/TableView';
import {CaptionTextSmall, Form} from '@pachyderm/components';

import RepositoriesList from './components/RepositoriesList';
import useRepoFilters, {repoFilters} from './hooks/useRepoFilters';
import useRepositoriesList from './hooks/useRepoList';
import styles from './ReposTable.module.css';

type ReposTableProps = {
  filtersExpanded: boolean;
};

const ReposTable: React.FC<ReposTableProps> = ({filtersExpanded}) => {
  const {
    repos,
    loading,
    error,
    pageIndex,
    updatePage,
    pageSize,
    setPageSize,
    hasNextPage,
  } = useRepositoriesList();
  const {sortedRepos, formCtx, staticFilterKeys} = useRepoFilters({repos});

  return (
    <Form formContext={formCtx} data-testid="ReposTable__table">
      {pageIndex > 0 || hasNextPage ? (
        filtersExpanded ? (
          <CaptionTextSmall className={styles.noFilters}>
            There are too many repos to enable sorting and filtering
          </CaptionTextSmall>
        ) : null
      ) : (
        <TableViewFilters
          formCtx={formCtx}
          filtersExpanded={filtersExpanded}
          filters={repoFilters}
          staticFilterKeys={staticFilterKeys}
        />
      )}
      <RepositoriesList
        loading={loading}
        error={error}
        repos={sortedRepos}
        totalReposLength={repos?.length || 0}
        pageIndex={pageIndex}
        updatePage={updatePage}
        pageSize={pageSize}
        setPageSize={setPageSize}
        hasNextPage={hasNextPage}
      />
    </Form>
  );
};

export default ReposTable;
