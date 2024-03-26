import React from 'react';

import {TableViewFilters} from '@dash-frontend/components/TableView';
import {CaptionTextSmall, Form} from '@pachyderm/components';

import PipelineStepsList from './components/PipelineStepsList';
import usePipelineFilters, {pipelineFilters} from './hooks/usePipelineFilters';
import usePipelinesList from './hooks/usePipelineList';
import styles from './PipelineStepsTable.module.css';

type PipelineStepsTableProps = {
  filtersExpanded: boolean;
};

const PipelineStepsTable: React.FC<PipelineStepsTableProps> = ({
  filtersExpanded,
}) => {
  const {
    pipelines,
    loading,
    error,
    pageIndex,
    updatePage,
    pageSize,
    setPageSize,
    hasNextPage,
    isAuthActive,
  } = usePipelinesList();
  const {sortedPipelines, formCtx, staticFilterKeys, clearableFiltersMap} =
    usePipelineFilters({pipelines});

  return (
    <Form formContext={formCtx} data-testid="PipelineStepsTable__table">
      {pageIndex > 0 || hasNextPage ? (
        filtersExpanded ? (
          <CaptionTextSmall className={styles.noFilters}>
            There are too many pipelines to enable sorting and filtering
          </CaptionTextSmall>
        ) : null
      ) : (
        <TableViewFilters
          formCtx={formCtx}
          filtersExpanded={filtersExpanded}
          filters={pipelineFilters}
          clearableFiltersMap={clearableFiltersMap}
          staticFilterKeys={staticFilterKeys}
        />
      )}
      <PipelineStepsList
        loading={loading}
        error={error}
        pipelines={sortedPipelines}
        totalPipelinesLength={pipelines?.length || 0}
        pageIndex={pageIndex}
        updatePage={updatePage}
        pageSize={pageSize}
        setPageSize={setPageSize}
        hasNextPage={hasNextPage}
        isAuthActive={isAuthActive}
      />
    </Form>
  );
};

export default PipelineStepsTable;
