import {useQuery} from '@tanstack/react-query';

import {Project} from '@dash-frontend/api/pfs';
import {listPipeline} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const usePipelines = (
  projectName: Project['name'],
  enabled = true,
  staleTime?: number,
) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.pipelines({projectId: projectName}),
    queryFn: () => {
      return listPipeline({projects: [{name: projectName}]});
    },
    enabled,
    staleTime,
  });

  return {
    loading,
    pipelines: data,
    error: getErrorMessage(error),
  };
};

export const usePaginatedPipelines = (
  projectName: Project['name'],
  pageSize: number,
  pageIndex: number,
) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.pipelines({
      projectId: projectName,
      pageIndex,
      pageSize,
    }),
    queryFn: () =>
      listPipeline({
        projects: [{name: projectName}],
        page: {pageSize: String(pageSize), pageIndex: String(pageIndex)},
      }),
  });

  return {
    loading,
    pipelines: data,
    error: getErrorMessage(error),
  };
};

export const usePipelinesLazy = (projectName: Project['name']) => {
  const {
    data,
    isLoading: loading,
    isFetched,
    error,
    refetch,
  } = useQuery({
    queryKey: queryKeys.pipelines({projectId: projectName}),
    queryFn: () => {
      return listPipeline({projects: [{name: projectName}]});
    },
    enabled: false,
  });

  return {
    getPipelines: refetch,
    loading,
    isFetched,
    pipelines: data,
    error: getErrorMessage(error),
  };
};
