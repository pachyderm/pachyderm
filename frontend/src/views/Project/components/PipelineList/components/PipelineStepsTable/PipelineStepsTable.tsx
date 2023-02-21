import React from 'react';

import {
  TableViewFilters,
  TableViewLoadingDots,
} from '@dash-frontend/components/TableView';
import usePipelines from '@dash-frontend/hooks/usePipelines';
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
  const {sortedPipelines, formCtx, staticFilterKeys, clearableFiltersMap} =
    usePipelineFilters({pipelines});

  return (
    <Form formContext={formCtx}>
      <TableViewFilters
        formCtx={formCtx}
        filtersExpanded={filtersExpanded}
        filters={pipelineFilters}
        clearableFiltersMap={clearableFiltersMap}
        staticFilterKeys={staticFilterKeys}
      />

      {loading ? (
        <TableViewLoadingDots data-testid="PipelineStepsTable__loadingDots" />
      ) : (
        <PipelineStepsList
          loading={loading}
          error={error}
          pipelines={sortedPipelines}
          totalPipelinesLength={pipelines?.length || 0}
        />
      )}
    </Form>
  );
};

export default PipelineStepsTable;
