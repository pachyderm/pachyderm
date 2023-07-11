import {ReposWithCommitQuery} from '@graphqlTypes';
import React, {useMemo} from 'react';

import {
  TableViewFilters,
  TableViewLoadingDots,
} from '@dash-frontend/components/TableView';
import usePipelines from '@dash-frontend/hooks/usePipelines';
import useReposWithLinkedPipeline from '@dash-frontend/hooks/useReposWithLinkedPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Form} from '@pachyderm/components';

import PipelineStepsList from './components/PipelineStepsList';
import usePipelineFilters, {pipelineFilters} from './hooks/usePipelineFilters';

type PipelineStepsTableProps = {
  filtersExpanded: boolean;
};

const PipelineStepsTable: React.FC<PipelineStepsTableProps> = ({
  filtersExpanded,
}) => {
  const {projectId} = useUrlState();
  const {pipelines, loading, error} = usePipelines({projectIds: [projectId]});
  const {
    repos,
    loading: reposLoading,
    error: reposError,
  } = useReposWithLinkedPipeline({projectId});
  const {sortedPipelines, formCtx, staticFilterKeys, clearableFiltersMap} =
    usePipelineFilters({pipelines});

  const pipelineRepoMap = useMemo(() => {
    const obj: Record<string, ReposWithCommitQuery['repos'][0]> = {};
    repos?.forEach((repo) => {
      if (repo?.linkedPipeline?.id) {
        obj[repo?.linkedPipeline?.id] = repo;
      }
    });
    return obj;
  }, [repos]);

  return (
    <Form formContext={formCtx}>
      <TableViewFilters
        formCtx={formCtx}
        filtersExpanded={filtersExpanded}
        filters={pipelineFilters}
        clearableFiltersMap={clearableFiltersMap}
        staticFilterKeys={staticFilterKeys}
      />

      {loading || reposLoading ? (
        <TableViewLoadingDots data-testid="PipelineStepsTable__loadingDots" />
      ) : (
        <PipelineStepsList
          loading={loading}
          error={error || reposError}
          pipelines={sortedPipelines}
          totalPipelinesLength={pipelines?.length || 0}
          pipelineRepoMap={pipelineRepoMap}
        />
      )}
    </Form>
  );
};

export default PipelineStepsTable;
