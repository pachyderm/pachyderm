import {DagQueryArgs} from '@graphqlTypes';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';

export const useDAGData = ({
  projectId,
  nodeHeight,
  nodeWidth,
  direction,
}: DagQueryArgs) => {
  const {data, error, loading} = useGetDagQuery({
    variables: {args: {projectId, nodeHeight, nodeWidth, direction}},
  });

  return {
    error,
    dag: data?.dag,
    loading,
  };
};
