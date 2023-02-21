import React, {useState} from 'react';

import TableView, {TableViewSection} from '@dash-frontend/components/TableView';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {PROJECT_PIPELINES_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

import JobsList from '../JobsList';

import PipelineStepsTable from './components/PipelineStepsTable';

const TAB_IDS = {steps: 'steps', jobs: 'jobs'};

const PipelineList: React.FC = () => {
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const {viewState} = useUrlQueryState();

  return (
    <TableView
      title="Pipeline Steps"
      noun="pipeline step"
      tabsBasePath={PROJECT_PIPELINES_PATH}
      tabs={TAB_IDS}
      selectedItems={viewState.selectedPipelines || []}
      filtersExpanded={filtersExpanded}
      setFiltersExpanded={setFiltersExpanded}
    >
      <TableViewSection id={TAB_IDS.steps}>
        <PipelineStepsTable filtersExpanded={filtersExpanded} />
      </TableViewSection>
      <TableViewSection id={TAB_IDS.jobs}>
        <JobsList
          selectedPipelines={viewState.selectedPipelines}
          filtersExpanded={filtersExpanded}
        />
      </TableViewSection>
    </TableView>
  );
};

export default PipelineList;
