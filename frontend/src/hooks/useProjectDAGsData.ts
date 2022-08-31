import {DagQueryArgs, GetDagsSubscription, NodeType} from '@graphqlTypes';
import {useEffect, useState} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import {useGetDagsSubscription} from '@dash-frontend/generated/hooks';
import buildDags from '@dash-frontend/lib/dag';
import deriveRepoNameFromNode from '@dash-frontend/lib/deriveRepoNameFromNode';
import {Dag, DagDirection} from '@dash-frontend/lib/types';
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

interface useProjectDagsDataProps extends DagQueryArgs {
  nodeWidth: number;
  nodeHeight: number;
  direction?: DagDirection;
}

interface buildAndSetDagsProps
  extends Omit<useProjectDagsDataProps, 'projectId'> {
  data?: GetDagsSubscription;
  projectMatch: boolean;
}

export const useProjectDagsData = ({
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
  const [prevData, setPrevData] = useState<GetDagsSubscription | undefined>();
  const [dags, setDags] = useState<Dag[] | undefined>();

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
  } = useGetDagsSubscription({
    variables: {args: {projectId, jobSetId}},
    onSubscriptionData: async ({client, subscriptionData}) => {
      if (
        (repoId || pipelineId) &&
        !(subscriptionData.data?.dags || []).some((dag) => {
          return (
            (dag.type === NodeType.PIPELINE && dag.name === pipelineId) ||
            (dag.type !== NodeType.PIPELINE &&
              deriveRepoNameFromNode(dag) === repoId)
          );
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
    projectMatch,
  }: buildAndSetDagsProps) => {
    setIsFirstCall(false);
    if (data?.dags) {
      const dags = await buildDags(
        data?.dags,
        nodeWidth,
        nodeHeight,
        direction,
        setDagError,
        projectMatch,
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
