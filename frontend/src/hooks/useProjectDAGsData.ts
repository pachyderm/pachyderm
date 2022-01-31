import {DagQueryArgs, GetDagsSubscription, NodeType} from '@graphqlTypes';
import isEqual from 'lodash/isEqual';
import {useEffect, useState} from 'react';
import {useHistory} from 'react-router';

import {useGetDagsSubscription} from '@dash-frontend/generated/hooks';
import buildDags from '@dash-frontend/lib/dag';
import deriveRepoNameFromNode from '@dash-frontend/lib/deriveRepoNameFromNode';
import {Dag, DagDirection} from '@dash-frontend/lib/types';
import {projectRoute} from '@dash-frontend/views/Project/utils/routes';

import useUrlState from './useUrlState';

interface useProjectDagsDataProps extends DagQueryArgs {
  nodeWidth: number;
  nodeHeight: number;
  direction?: DagDirection;
}

interface buildAndSetDagsProps
  extends Omit<useProjectDagsDataProps, 'projectId'> {
  data?: GetDagsSubscription;
}

export const useProjectDagsData = ({
  jobSetId,
  projectId,
  nodeWidth,
  nodeHeight,
  direction = DagDirection.RIGHT,
}: useProjectDagsDataProps) => {
  const {repoId, pipelineId, projectId: routeProjectId} = useUrlState();
  const browserHistory = useHistory();
  const [isFirstCall, setIsFirstCall] = useState(true);
  const [isLoading, setIsLoading] = useState(true);
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
        browserHistory.push(projectRoute({projectId}));
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
    jobSetId,
  }: buildAndSetDagsProps) => {
    setIsFirstCall(false);
    if (data?.dags) {
      const dags = await buildDags(
        data?.dags,
        nodeWidth,
        nodeHeight,
        direction,
        jobSetId ? jobSetId : undefined,
      );
      setDags(dags);
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (!error) {
      setPrevData(data);
    } else setIsLoading(false);
  }, [error, data]);

  useEffect(() => {
    // only rebuild the dag when subscription data has changed, and not jobSetId
    if (!isEqual(data, prevData)) {
      buildAndSetDags({data, nodeWidth, nodeHeight, direction, jobSetId});
    }
  }, [data, nodeHeight, nodeWidth, direction, prevData, jobSetId]);

  return {
    error,
    dags,
    loading: isLoading || isSubscribing,
  };
};
