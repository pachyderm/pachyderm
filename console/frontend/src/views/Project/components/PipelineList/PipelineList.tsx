import React, {useState} from 'react';

import TableView, {TableViewSection} from '@dash-frontend/components/TableView';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {PROJECT_PIPELINES_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

import JobsList from '../JobsList';

import PipelinesRuntimeChart from './components/PipelinesRuntimeChart';
import PipelineStepsTable from './components/PipelineStepsTable';

const TAB_IDS = {
  pipelines: 'pipelines',
  jobs: 'jobs',
  runtimes: 'runtimes',
};

const PipelineList: React.FC = () => {
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const {searchParams} = useUrlQueryState();

  return (
    <TableView
      title="Pipelines"
      noun="pipeline"
      tabsBasePath={PROJECT_PIPELINES_PATH}
      tabs={TAB_IDS}
      selectedItems={searchParams.selectedPipelines || []}
      filtersExpanded={filtersExpanded}
      setFiltersExpanded={setFiltersExpanded}
    >
      <TableViewSection id={TAB_IDS.pipelines}>
        <PipelineStepsTable filtersExpanded={filtersExpanded} />
      </TableViewSection>
      <TableViewSection id={TAB_IDS.jobs}>
        <JobsList
          selectedPipelines={searchParams.selectedPipelines}
          filtersExpanded={filtersExpanded}
        />
      </TableViewSection>
      <TableViewSection id={TAB_IDS.runtimes}>
        <PipelinesRuntimeChart
          selectedPipelines={searchParams.selectedPipelines}
          filtersExpanded={filtersExpanded}
        />
      </TableViewSection>
    </TableView>
  );
};

export default PipelineList;
