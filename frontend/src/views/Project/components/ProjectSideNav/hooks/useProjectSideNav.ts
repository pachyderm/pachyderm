import {JobState} from '@graphqlTypes';
import {useMemo, useCallback} from 'react';
import {useRouteMatch} from 'react-router';

import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  lineageRoute,
  jobsRoute,
} from '@dash-frontend/views/Project/utils/routes';

import {
  LINEAGE_PATH,
  PROJECT_JOBS_PATH,
  LINEAGE_JOBS_PATH,
} from '../../../constants/projectPaths';

export const useProjectSideNav = () => {
  const {projectId} = useUrlState();
  const {jobSets} = useJobSets({projectId});
  const [, handleListDefaultView] = useLocalProjectSettings({
    projectId,
    key: 'list_view_default',
  });

  const numOfFailedJobs = useMemo(() => {
    return jobSets.length > 0
      ? jobSets[0].jobs.filter((job) => job.state === JobState.JOB_FAILURE)
          .length
      : 0;
  }, [jobSets]);

  const jobsMatch = useRouteMatch({
    path: [PROJECT_JOBS_PATH, LINEAGE_JOBS_PATH],
    exact: true,
  });
  const lineageMatch = useRouteMatch({
    path: LINEAGE_PATH,
  });

  const getJobsLink = useCallback(() => {
    const jobsClosedPath = lineageMatch
      ? lineageRoute({projectId})
      : jobsRoute({projectId});

    return jobsMatch ? jobsClosedPath : jobsRoute({projectId});
  }, [jobsMatch, lineageMatch, projectId]);

  return {
    projectId,
    numOfFailedJobs,
    handleListDefaultView,
    getJobsLink,
  };
};
