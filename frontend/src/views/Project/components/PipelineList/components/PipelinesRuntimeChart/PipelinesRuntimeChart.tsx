import React, {useState} from 'react';

import {useJobsByPipeline} from '@dash-frontend/hooks/useJobs';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import RuntimesChart from '@dash-frontend/views/Project/components/RuntimesChart';
import {ChevronDownSVG, Dropdown} from '@pachyderm/components';

type PipelinesRuntimeChartProps = {
  filtersExpanded: boolean;
  selectedPipelines?: string[];
};

const LIMIT_OPTIONS = [3, 6, 10];

const PipelinesRuntimeChart: React.FC<PipelinesRuntimeChartProps> = ({
  filtersExpanded,
  selectedPipelines,
}) => {
  const {projectId} = useUrlState();
  const [limit, setLimit] = useState(LIMIT_OPTIONS[0]);

  const {data, error, loading} = useJobsByPipeline({
    limit,
    project: projectId,
    pipelineIds: selectedPipelines || [],
  });

  const viewOptions = (
    <Dropdown>
      <Dropdown.Button IconSVG={ChevronDownSVG} buttonType="dropdown">
        {`Last ${limit} Jobs`}
      </Dropdown.Button>
      <Dropdown.Menu>
        {LIMIT_OPTIONS.map((limitOption) => (
          <Dropdown.MenuItem
            key={`jobsLimit${limitOption}`}
            id={`jobsLimit${limitOption}`}
            onClick={() => setLimit(limitOption)}
            closeOnClick
          >
            {`Last ${limitOption} Jobs`}
          </Dropdown.MenuItem>
        ))}
      </Dropdown.Menu>
    </Dropdown>
  );

  return (
    <RuntimesChart
      jobs={data}
      loading={loading}
      error={error}
      viewOptions={viewOptions}
      resource="pipeline"
      filtersExpanded={filtersExpanded}
    />
  );
};

export default PipelinesRuntimeChart;
