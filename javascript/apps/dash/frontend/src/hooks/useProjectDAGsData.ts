import {useGetDagsQuery} from '@dash-frontend/generated/hooks';
import {DagQueryArgs} from '@graphqlTypes';

export const useProjectDagsData = ({
  projectId,
  nodeWidth,
  nodeHeight,
}: DagQueryArgs) => {
  const {data, error, loading} = useGetDagsQuery({
    variables: {args: {projectId, nodeHeight, nodeWidth}},
  });

  return {
    error,
    dags: data?.dags,
    loading,
  };
};
