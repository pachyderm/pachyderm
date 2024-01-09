import React from 'react';

import {
  TableViewFilters,
  TableViewLoadingDots,
} from '@dash-frontend/components/TableView';
import {useRepos} from '@dash-frontend/hooks/useRepos';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Form} from '@pachyderm/components';

import RepositoriesList from './components/RepositoriesList';
import useRepoFilters, {repoFilters} from './hooks/useRepoFilters';

type ReposTableProps = {
  filtersExpanded: boolean;
};

const ReposTable: React.FC<ReposTableProps> = ({filtersExpanded}) => {
  const {projectId} = useUrlState();
  const {repos, loading, error} = useRepos(projectId);
  const {sortedRepos, formCtx, staticFilterKeys} = useRepoFilters({repos});

  return (
    <Form formContext={formCtx}>
      <TableViewFilters
        formCtx={formCtx}
        filtersExpanded={filtersExpanded}
        filters={repoFilters}
        staticFilterKeys={staticFilterKeys}
      />
      {loading ? (
        <TableViewLoadingDots data-testid="ReposTable__loadingDots" />
      ) : (
        <RepositoriesList
          loading={loading}
          error={error}
          repos={sortedRepos}
          totalReposLength={repos?.length || 0}
        />
      )}
    </Form>
  );
};

export default ReposTable;
