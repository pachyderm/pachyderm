import {useEffect, useState, useCallback} from 'react';

import {usePipelinesLazy} from '@dash-frontend/hooks/usePipelines';
import {useReposLazy} from '@dash-frontend/hooks/useRepos';
import {
  convertPachydermTypesToVertex,
  Vertex,
} from '@dash-frontend/lib/DAG/convertPipelinesAndReposToVertices';

import useUrlState from './useUrlState';

export const useVerticesLazy = () => {
  const {projectId} = useUrlState();

  const [vertices, setVertices] = useState<Vertex[]>();
  const [isConverting, setIsConverting] = useState(true);

  const {
    getRepos,
    repos,
    loading: reposLoading,
    error: reposError,
    isFetched: isReposFetched,
  } = useReposLazy(projectId);
  const {
    getPipelines,
    pipelines,
    loading: pipelinesLoading,
    error: pipelinesError,
    isFetched: isPipelinesFetched,
  } = usePipelinesLazy(projectId);

  const error = reposError || pipelinesError;
  const loading = reposLoading || pipelinesLoading;

  const convertVertices = useCallback(async () => {
    if (!loading && isPipelinesFetched && isReposFetched) {
      const vertices = await convertPachydermTypesToVertex(
        projectId,
        repos,
        pipelines,
      );
      setIsConverting(false);
      setVertices(vertices);
    }
  }, [
    isPipelinesFetched,
    isReposFetched,
    loading,
    pipelines,
    projectId,
    repos,
  ]);

  useEffect(() => {
    convertVertices();
  }, [convertVertices]);

  const getVertices = () => {
    getRepos();
    getPipelines();
  };

  return {
    getVertices,
    error,
    vertices,
    loading: isConverting,
  };
};
