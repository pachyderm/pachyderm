import cloneDeep from 'lodash/cloneDeep';
import countBy from 'lodash/countBy';
import {useMemo} from 'react';
import {useHistory} from 'react-router';

import {PipelineJobFilters} from '@dash-frontend/components/JobList/hooks/useJobList';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {PipelineJobsQuery, PipelineJobState} from '@graphqlTypes';

const useJobListStatusFilter = (
  pipelineJobs: PipelineJobsQuery['pipelineJobs'],
  selectedFilters: PipelineJobFilters,
) => {
  const {setUrlFromViewState} = useUrlQueryState();
  const browserHistory = useHistory();

  const stateCounts = useMemo(
    () => countBy(pipelineJobs, (job) => job.state),
    [pipelineJobs],
  );

  const onChipClick = (job?: PipelineJobState) => {
    if (job) {
      const updatedFilterList = cloneDeep(selectedFilters);
      updatedFilterList[job] = !updatedFilterList[job];

      const updatedPipelineJobFilters = Object.entries(
        updatedFilterList,
      ).reduce<PipelineJobState[]>((result, [key, selected]) => {
        if (selected) {
          result.push(key as PipelineJobState);
        }
        return result;
      }, []);
      browserHistory.push(
        setUrlFromViewState({pipelineJobFilters: updatedPipelineJobFilters}),
      );
    }
  };

  return {stateCounts, onChipClick};
};
export default useJobListStatusFilter;
