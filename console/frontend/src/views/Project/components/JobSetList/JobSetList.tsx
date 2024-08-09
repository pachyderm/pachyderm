import React, {useState} from 'react';

import TableView, {TableViewSection} from '@dash-frontend/components/TableView';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {PROJECT_JOBS_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

import JobsList from '../JobsList';

import JobsRuntimeChart from './components/JobsRuntimeChart';
import RunsTable from './components/RunsTable';

const TAB_IDS = {jobs: 'jobs', subjobs: 'subjobs', runtimes: 'runtimes'};

const JobSetList: React.FC = () => {
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const {searchParams} = useUrlQueryState();

  return (
    <TableView
      title="Jobs"
      noun="job"
      tabsBasePath={PROJECT_JOBS_PATH}
      tabs={TAB_IDS}
      selectedItems={searchParams.selectedJobs || []}
      filtersExpanded={filtersExpanded}
      setFiltersExpanded={setFiltersExpanded}
    >
      <TableViewSection id={TAB_IDS.jobs}>
        <RunsTable filtersExpanded={filtersExpanded} />
      </TableViewSection>
      <TableViewSection id={TAB_IDS.subjobs}>
        <JobsList
          selectedJobSets={searchParams.selectedJobs}
          filtersExpanded={filtersExpanded}
        />
      </TableViewSection>
      <TableViewSection id={TAB_IDS.runtimes}>
        <JobsRuntimeChart
          selectedJobSets={searchParams.selectedJobs}
          filtersExpanded={filtersExpanded}
        />
      </TableViewSection>
    </TableView>
  );
};

export default JobSetList;
