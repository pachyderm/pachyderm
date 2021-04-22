import {useGetDagsSubscription} from '@dash-frontend/generated/hooks';
import {DagQueryArgs} from '@graphqlTypes';

export const useProjectDagsData = ({
  projectId,
  nodeWidth,
  nodeHeight,
}: DagQueryArgs) => {
  const {data, error, loading} = useGetDagsSubscription({
    variables: {args: {projectId, nodeHeight, nodeWidth}},
  });

  return {
    error,
    dags: data?.dags,
    loading,
  };
};
