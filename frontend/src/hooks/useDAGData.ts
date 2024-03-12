import {useQueryClient} from '@tanstack/react-query';
import {useEffect, useState, useCallback} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import {ProjectStatus} from '@dash-frontend/api/pfs';
import {useJobSet} from '@dash-frontend/hooks/useJobSet';
import {usePipelines} from '@dash-frontend/hooks/usePipelines';
import {useRepos} from '@dash-frontend/hooks/useRepos';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import buildELKDags from '@dash-frontend/lib/DAG/buildELKDags';
import {getProjectGlobalIdVertices} from '@dash-frontend/lib/DAG/convertJobsToVertices';
import {
  convertPachydermTypesToVertex,
  Vertex,
} from '@dash-frontend/lib/DAG/convertPipelinesAndReposToVertices';
import {sortJobInfos} from '@dash-frontend/lib/DAG/DAGhelpers';
import queryKeys from '@dash-frontend/lib/queryKeys';
import {Dags, DagDirection} from '@dash-frontend/lib/types';
import {
  PROJECT_PATH,
  PROJECT_PIPELINES_PATH,
  PROJECT_REPOS_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  lineageRoute,
  projectReposRoute,
  projectPipelinesRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {usePreviousValue} from '@pachyderm/components';

import {deriveProjectStatus} from './useProjectStatus';
import useUrlState from './useUrlState';

interface useProjectDagsDataProps {
  direction?: DagDirection;
}

export const useDAGData = ({
  direction = DagDirection.DOWN,
}: useProjectDagsDataProps) => {
  const {searchParams} = useUrlQueryState();
  const globalId = searchParams.globalIdFilter;
  const {repoId, pipelineId, projectId} = useUrlState();
  const browserHistory = useHistory();
  const projectMatch = !!useRouteMatch(PROJECT_PATH);
  const projectReposMatch = useRouteMatch(PROJECT_REPOS_PATH);
  const projectPipelinesMatch = useRouteMatch(PROJECT_PIPELINES_PATH);
  const [dagError, setDagError] = useState<string>();
  const [dags, setDags] = useState<Dags | undefined>();
  const [isBuilding, setIsBuilding] = useState(true);

  const {
    repos,
    loading: reposLoading,
    error: reposError,
  } = useRepos(projectId, !globalId);
  const {
    pipelines,
    loading: pipelinesLoading,
    error: pipelinesError,
  } = usePipelines(projectId, !globalId);

  // This code is used to invalidate the pipelines and projectStatus queries
  // so that the search bar will reflect any state changes in the DAG.
  const client = useQueryClient();
  const [projectStatus, setProjectStatus] = useState<ProjectStatus>(
    ProjectStatus.HEALTHY,
  );
  const previousProjectStatus = usePreviousValue<ProjectStatus>(projectStatus);

  useEffect(() => {
    setProjectStatus(deriveProjectStatus(pipelines));
  }, [pipelines]);

  useEffect(() => {
    if (
      projectStatus &&
      previousProjectStatus &&
      projectStatus !== previousProjectStatus
    ) {
      client.invalidateQueries({
        queryKey: queryKeys.projectStatus({projectId: projectId}),
      });
      client.invalidateQueries({
        queryKey: queryKeys.pipelines({projectId: projectId}),
      });
    }
  }, [client, previousProjectStatus, projectId, projectStatus]);

  const {
    jobSet,
    loading: jobSetLoading,
    error: jobSetError,
  } = useJobSet(globalId, Boolean(globalId));

  const error = reposError || pipelinesError || jobSetError;
  const loading = reposLoading || pipelinesLoading || jobSetLoading;

  const buildAndSetDags = useCallback(async () => {
    if (!loading) {
      let vertices: Vertex[];
      // if given a global ID, we want to construct a DAG using only
      // nodes referred to by jobs sharing that global id
      // stabilize elkjs inputs.
      if (globalId) {
        // listRepo and listPipeline are both stabilized on 'name',
        // but jobsets come back in a stream with random order.
        const sortedJobSet = sortJobInfos(jobSet || []);
        vertices = await getProjectGlobalIdVertices(sortedJobSet, projectId);
      } else {
        vertices = await convertPachydermTypesToVertex(
          projectId,
          repos,
          pipelines,
        );
      }
      const dags = await buildELKDags(vertices, direction, setDagError);
      setDags(dags);
      setIsBuilding(false);
    }
  }, [direction, globalId, jobSet, loading, pipelines, projectId, repos]);

  useEffect(() => {
    buildAndSetDags();
  }, [buildAndSetDags]);

  useEffect(() => {
    if (
      !isBuilding &&
      (repoId || pipelineId) &&
      !(dags?.nodes || []).some(({name}) => {
        return name === pipelineId || name === repoId;
      })
    ) {
      // redirect if we were at a route that got deleted or filtered out
      // by a global id filter and is no longer in any dags
      if (projectMatch) {
        if (projectReposMatch)
          browserHistory.push(projectReposRoute({projectId}));
        if (projectPipelinesMatch)
          browserHistory.push(projectPipelinesRoute({projectId}));
      } else {
        browserHistory.push(lineageRoute({projectId}));
      }
    }
  }, [
    browserHistory,
    dags,
    isBuilding,
    loading,
    pipelineId,
    projectId,
    projectMatch,
    projectPipelinesMatch,
    projectReposMatch,
    repoId,
  ]);

  return {
    error: error || dagError,
    dags,
    loading: isBuilding,
  };
};
