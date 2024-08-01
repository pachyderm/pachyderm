import React from 'react';

import {useJobs} from '@dash-frontend/hooks/useJobs';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import RuntimesChart from '@dash-frontend/views/Project/components/RuntimesChart';

type JobsRuntimeChartProps = {
  filtersExpanded: boolean;
  selectedJobSets?: string[];
};

const JobsRuntimeChart: React.FC<JobsRuntimeChartProps> = ({
  filtersExpanded,
  selectedJobSets,
}) => {
  const {projectId} = useUrlState();
  const {jobs, error, loading} = useJobs({
    projectName: projectId,
    jobSetIds: selectedJobSets,
  });

  return (
    <RuntimesChart
      jobs={jobs}
      loading={loading}
      error={error}
      resource="job"
      filtersExpanded={filtersExpanded}
    />
  );
};

export default JobsRuntimeChart;
