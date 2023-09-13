import {VerticesQueryArgs, VerticesSubscription} from '@graphqlTypes';
import {useEffect, useState} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import {useVerticesSubscription} from '@dash-frontend/generated/hooks';
import buildDags from '@dash-frontend/lib/dag';
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

import useUrlState from './useUrlState';

interface useProjectDagsDataProps extends VerticesQueryArgs {
  nodeWidth: number;
  nodeHeight: number;
  direction?: DagDirection;
}

interface buildAndSetDagsProps
  extends Omit<useProjectDagsDataProps, 'projectId'> {
  data?: VerticesSubscription;
  projectMatch: boolean;
}

export const useGenerateDagsSubscription = ({
  jobSetId,
  projectId,
  nodeWidth,
  nodeHeight,
  direction = DagDirection.DOWN,
}: useProjectDagsDataProps) => {
  const {repoId, pipelineId, projectId: routeProjectId} = useUrlState();
  const browserHistory = useHistory();
  const projectMatch = !!useRouteMatch(PROJECT_PATH);
  const projectReposMatch = useRouteMatch(PROJECT_REPOS_PATH);
  const projectPipelinesMatch = useRouteMatch(PROJECT_PIPELINES_PATH);
  const [isFirstCall, setIsFirstCall] = useState(true);
  const [isLoading, setIsLoading] = useState(true);
  const [dagError, setDagError] = useState<string>();
  const [prevData, setPrevData] = useState<VerticesSubscription | undefined>();
  const [dags, setDags] = useState<Dags | undefined>();

  useEffect(() => {
    setIsLoading(true);
  }, [jobSetId]);

  useEffect(() => {
    if (isFirstCall) setIsLoading(true);
  }, [isFirstCall]);

  const {
    data = prevData,
    error,
    loading: isSubscribing,
  } = useVerticesSubscription({
    variables: {args: {projectId, jobSetId}},
    onSubscriptionData: async ({client, subscriptionData}) => {
      if (
        (repoId || pipelineId) &&
        !(subscriptionData.data?.vertices || []).some(({name}) => {
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
      } else {
        client.reFetchObservableQueries();
      }
    },
    // We want to skip the `onSubscriptionData` callback if
    // we've been redirected to, for example, the NOT_FOUND page
    // and the consuming component has been unmounted
    skip: !routeProjectId,
  });

  const buildAndSetDags = async ({
    data,
    nodeWidth,
    nodeHeight,
    direction,
  }: buildAndSetDagsProps) => {
    setIsFirstCall(false);
    if (data?.vertices) {
      const dags = await buildDags(
        data?.vertices,
        nodeWidth,
        nodeHeight,
        direction,
        setDagError,
      );
      setDags(dags);
      setIsLoading(false);
    }
  };

  useEffect(() => {
    setIsLoading(true);
  }, [projectMatch]);

  useEffect(() => {
    if (!error) {
      setPrevData(data);
    } else setIsLoading(false);
  }, [error, data]);

  useEffect(() => {
    buildAndSetDags({
      data,
      nodeWidth,
      nodeHeight,
      direction,
      projectMatch,
    });
  }, [data, nodeHeight, nodeWidth, direction, projectMatch]);

  return {
    error: error || dagError,
    dags,
    loading: isLoading || isSubscribing,
  };
};
