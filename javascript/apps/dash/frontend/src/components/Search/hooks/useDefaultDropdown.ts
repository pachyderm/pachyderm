import countBy from 'lodash/countBy';
import sum from 'lodash/sum';
import values from 'lodash/values';
import {useMemo} from 'react';
import {useHistory} from 'react-router';

import useJobList from '@dash-frontend/components/JobList/hooks/useJobList';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {jobsRoute} from '@dash-frontend/views/Project/utils/routes';

export const useDefaultDropdown = () => {
  const {projectId} = useUrlState();
  const {jobs} = useJobList({projectId});
  const browserHistory = useHistory();

  const stateCounts = useMemo(() => countBy(jobs, (job) => job.state), [jobs]);
  const allJobs = useMemo(() => sum(values(stateCounts)), [stateCounts]);

  const routeToJobs = () => {
    browserHistory.push(jobsRoute({projectId}));
  };

  return {stateCounts, allJobs, routeToJobs};
};
