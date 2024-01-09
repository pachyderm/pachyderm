import React from 'react';

import {
  TableViewFilters,
  TableViewLoadingDots,
} from '@dash-frontend/components/TableView';
import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Form} from '@pachyderm/components';

import RunsList from './components/RunsList';
import useJobSetFilters, {jobSetFilters} from './hooks/useJobsetFilters';

type RunsTableProps = {
  filtersExpanded: boolean;
};

const RunsTable: React.FC<RunsTableProps> = ({filtersExpanded}) => {
  const {projectId} = useUrlState();
  const {
    jobs: jobSets,
    loading: jobSetsLoading,
    error: jobSetsError,
  } = useJobSets(projectId);
  const {sortedJobsets, formCtx, staticFilterKeys, clearableFiltersMap} =
    useJobSetFilters({jobSets});

  return (
    <Form formContext={formCtx}>
      <TableViewFilters
        formCtx={formCtx}
        filtersExpanded={filtersExpanded}
        filters={jobSetFilters}
        clearableFiltersMap={clearableFiltersMap}
        staticFilterKeys={staticFilterKeys}
      />
      {jobSetsLoading ? (
        <TableViewLoadingDots data-testid="RunsTable__loadingDots" />
      ) : (
        <RunsList
          error={jobSetsError}
          jobSets={sortedJobsets}
          totalJobsetsLength={jobSets?.length || 0}
        />
      )}
    </Form>
  );
};

export default RunsTable;
