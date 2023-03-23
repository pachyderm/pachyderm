import React from 'react';

import {useJobsQuery} from '@dash-frontend/generated/hooks';
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
  const {data, error, loading} = useJobsQuery({
    variables: {
      args: {
        projectId,
        jobSetIds: selectedJobSets,
      },
    },
  });

  return (
    <RuntimesChart
      jobs={data?.jobs}
      loading={loading}
      error={error}
      resource="job"
      filtersExpanded={filtersExpanded}
    />
  );
};

export default JobsRuntimeChart;
