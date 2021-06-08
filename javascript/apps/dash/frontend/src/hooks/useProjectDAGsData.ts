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
  const {repoId, pipelineId} = useUrlState();
  const browserHistory = useHistory();

  const {data, error, loading} = useGetDagsSubscription({
    variables: {args: {projectId, nodeHeight, nodeWidth, direction}},
    onSubscriptionData: ({client, subscriptionData}) => {
      if (
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
  });

  return {
    error,
    dags: data?.dags,
    loading,
  };
};
