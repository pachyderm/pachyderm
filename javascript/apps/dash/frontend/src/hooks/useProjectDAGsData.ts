import {useHistory} from 'react-router';

import {useGetDagsSubscription} from '@dash-frontend/generated/hooks';
import deriveRepoNameFromNode from '@dash-frontend/lib/deriveRepoNameFromNode';
import {projectRoute} from '@dash-frontend/views/Project/utils/routes';
import {DagQueryArgs, NodeType} from '@graphqlTypes';

import useUrlState from './useUrlState';

export const useProjectDagsData = ({
  projectId,
  nodeWidth,
  nodeHeight,
  direction,
}: DagQueryArgs) => {
  const {repoId, pipelineId, projectId: routeProjectId} = useUrlState();
  const browserHistory = useHistory();

  const {data, error, loading} = useGetDagsSubscription({
    variables: {args: {projectId, nodeHeight, nodeWidth, direction}},
    onSubscriptionData: ({client, subscriptionData}) => {
      if (
        (repoId || pipelineId) &&
        !(subscriptionData.data?.dags || []).some((dag) => {
          return dag.nodes.some(
            (node) =>
              (node.type === NodeType.PIPELINE && node.name === pipelineId) ||
              (node.type !== NodeType.PIPELINE &&
                deriveRepoNameFromNode(node) === repoId),
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

  return {
    error,
    dags: data?.dags,
    loading,
  };
};
