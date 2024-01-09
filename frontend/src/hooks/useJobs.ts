import {useQueries, useQuery} from '@tanstack/react-query';

import {Project} from '@dash-frontend/api/pfs';
import {JobInfo, listJob} from '@dash-frontend/api/pps';
import {jqCombine, jqIn, jqSelect} from '@dash-frontend/api/utils/jqHelpers';
import {DEFAULT_JOBS_LIMIT} from '@dash-frontend/constants/limits';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export type useJobsArgs = {
  projectName: Project['name'];
  pipelineIds?: string[];
  jobSetIds?: string[];
  limit?: number;
};

export const useJobs = (req: useJobsArgs, enabled = true) => {
  const {projectName, pipelineIds, jobSetIds, limit = DEFAULT_JOBS_LIMIT} = req;
  let jqFilter = '';

  if (jobSetIds && jobSetIds.length > 0) {
    jqFilter = jqCombine(jqFilter, jqSelect(jqIn('.job.id', jobSetIds)));
  }

  if (pipelineIds && pipelineIds.length > 0) {
    jqFilter = jqCombine(
      jqFilter,
      jqSelect(jqIn('.job.pipeline.name', pipelineIds)),
    );
  }

  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.jobs<useJobsArgs>({
      projectId: projectName,
      args: req,
    }),
    queryFn: () => {
      if (
        jobSetIds &&
        jobSetIds.length > 0 &&
        pipelineIds &&
        pipelineIds.length > 0
      ) {
        throw new Error('Cannot filter by both pipelineIds and jobSetIds');
      }

      return listJob({
        projects: [{name: projectName}],
        history: '-1',
        details: true,
        number: String(limit),
        jqFilter,
      });
    },
    enabled,
  });

  return {
    loading,
    jobs: data,
    error: getErrorMessage(error),
  };
};

export const useJobsByPipeline = ({
  pipelineIds,
  limit,
  project,
}: {
  pipelineIds: string[];
  project: string;
  limit?: number;
}) => {
  const number = limit || DEFAULT_JOBS_LIMIT;

  const queries = useQueries({
    queries: pipelineIds.map((pipelineId) => {
      return {
        queryKey: queryKeys.jobsByPipeline({
          projectId: project,
          pipelineId,
          args: {
            limit: number,
          },
        }),
        queryFn: () =>
          listJob({
            pipeline: {name: pipelineId, project: {name: project}},
            number: String(number),
          }),
      };
    }),
  });

  const data: JobInfo[] = queries
    .map((query) => query.data)
    .filter((data): data is JobInfo[] => typeof data !== 'undefined')
    .flat();
  const loading = queries.some((query) => query.isLoading);
  const error = queries.find((query) => !!query.error)?.error;

  return {data, error: getErrorMessage(error), loading};
};
