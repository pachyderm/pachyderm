import {useQuery} from '@tanstack/react-query';

import {Project, ProjectStatus} from '@dash-frontend/api/pfs';
import {pipelinesSummary, PipelinesSummary} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

const toNum = (val?: string) => Number(val || 0);

export const getProjectStatus = (pipelineSummary?: PipelinesSummary) => {
  return toNum(pipelineSummary?.failedPipelines) > 0 ||
    toNum(pipelineSummary?.unhealthyPipelines) > 0
    ? ProjectStatus.UNHEALTHY
    : ProjectStatus.HEALTHY;
};

export const usePipelineSummaries = (
  projectNames: Project['name'][],
  enabled = true,
) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.pipelinesSummary({projectIds: projectNames}),
    queryFn: () =>
      pipelinesSummary({
        projects: projectNames.map((name) => ({name})),
      }),
    enabled,
  });

  // if a project has no pipelines, no summary is returned from pachd
  const pipelineSummaries = projectNames.map((projectName) => {
    const summary = data?.summaries?.find(
      (summary) => summary.project?.name === projectName,
    );
    if (summary) {
      return {
        summary,
        projectStatus: getProjectStatus(summary),
        pipelineCount:
          toNum(summary?.activePipelines) +
          toNum(summary?.failedPipelines) +
          toNum(summary?.pausedPipelines),
      };
    } else {
      return {
        summary: {
          project: {name: projectName},
        },
        projectStatus: ProjectStatus.HEALTHY,
        pipelineCount: 0,
      };
    }
  });

  return {
    loading,
    pipelineSummaries,
    error: getErrorMessage(error),
  };
};

export const usePipelineSummary = (
  projectName: Project['name'],
  enabled = true,
) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.pipelinesSummary({projectIds: [projectName]}),
    queryFn: () =>
      pipelinesSummary({
        projects: [{name: projectName}],
      }),
    enabled,
  });

  const pipelineSummary = data?.summaries?.find(
    (summary) => summary.project?.name === projectName,
  );

  const projectStatus = getProjectStatus(pipelineSummary);
  const pipelineCount =
    toNum(pipelineSummary?.activePipelines) +
    toNum(pipelineSummary?.failedPipelines) +
    toNum(pipelineSummary?.pausedPipelines);

  return {
    loading,
    pipelineSummary,
    projectStatus,
    pipelineCount,
    error: getErrorMessage(error),
  };
};
