import React, {useState} from 'react';

import TableView, {TableViewSection} from '@dash-frontend/components/TableView';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {PROJECT_JOBS_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

import JobsList from '../JobsList';

import RunsTable from './components/RunsTable';

const TAB_IDS = {jobs: 'jobs', subjobs: 'subjobs'};

const JobSetList: React.FC = () => {
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const {viewState} = useUrlQueryState();

  return (
    <TableView
      title="Jobs"
      noun="job"
      tabsBasePath={PROJECT_JOBS_PATH}
      tabs={TAB_IDS}
      selectedItems={viewState.selectedJobs || []}
      filtersExpanded={filtersExpanded}
      setFiltersExpanded={setFiltersExpanded}
    >
      <TableViewSection id={TAB_IDS.jobs}>
        <RunsTable filtersExpanded={filtersExpanded} />
      </TableViewSection>
      <TableViewSection id={TAB_IDS.subjobs}>
        <JobsList
          selectedJobSets={viewState.selectedJobs}
          filtersExpanded={filtersExpanded}
        />
      </TableViewSection>
    </TableView>
  );
};

export default JobSetList;
