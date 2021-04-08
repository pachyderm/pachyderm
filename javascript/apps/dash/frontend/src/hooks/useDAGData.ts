import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import {DagQueryArgs} from '@graphqlTypes';

export const useDAGData = ({
  projectId,
  nodeHeight,
  nodeWidth,
}: DagQueryArgs) => {
  const {data, error, loading} = useGetDagQuery({
    variables: {args: {projectId, nodeHeight, nodeWidth}},
  });

  return {
    error,
    dag: data?.dag,
    loading,
  };
};
