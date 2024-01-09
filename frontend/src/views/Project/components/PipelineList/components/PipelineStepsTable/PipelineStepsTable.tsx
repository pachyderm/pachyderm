import React, {useMemo} from 'react';

import {Pipeline} from '@dash-frontend/api/pps';
import {
  TableViewFilters,
  TableViewLoadingDots,
} from '@dash-frontend/components/TableView';
import {usePipelines} from '@dash-frontend/hooks/usePipelines';
import {useRepos} from '@dash-frontend/hooks/useRepos';
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
  const {pipelines, loading, error} = usePipelines(projectId);
  const {repos, loading: reposLoading, error: reposError} = useRepos(projectId);
  const {sortedPipelines, formCtx, staticFilterKeys, clearableFiltersMap} =
    usePipelineFilters({pipelines});

  const pipelineRepoMap =
    useMemo(() => {
      if (pipelines && repos) {
        const pipelineMap = pipelines.reduce<Record<string, Pipeline>>(
          (map, pipeline) => {
            const pipelineId = pipeline?.pipeline?.name;
            if (map && pipeline && pipelineId) {
              map[pipelineId] = pipeline?.pipeline || {};
            }
            return map;
          },
          {},
        );

        // pipeline id -> repo
        return repos.reduce<Record<string, (typeof repos)[0]>>((obj, repo) => {
          const linkedPipeline =
            pipelineMap && pipelineMap[repo?.repo?.name || ''];
          if (linkedPipeline) {
            obj[linkedPipeline.name || ''] = repo;
          }
          return obj;
        }, {});
      }
    }, [repos, pipelines]) || {};

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
