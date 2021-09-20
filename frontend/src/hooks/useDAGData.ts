import {DagQueryArgs} from '@graphqlTypes';
import {useState} from 'react';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import buildDags from '@dash-frontend/lib/dag';
import {Dag, DagDirection} from '@dash-frontend/lib/types';

interface GetDagQueryProps extends DagQueryArgs {
  nodeHeight: number;
  nodeWidth: number;
  direction?: DagDirection;
}

export const useDAGData = ({
  projectId,
  nodeHeight,
  nodeWidth,
  direction = DagDirection.RIGHT,
}: GetDagQueryProps) => {
  const [loadingDags, setLoadingDags] = useState(true);
  const [dags, setDags] = useState<Dag[] | undefined>();
  const {error, loading} = useGetDagQuery({
    variables: {args: {projectId}},
    onCompleted: async (data) => {
      const builtDags = await buildDags(
        data.dag,
        nodeWidth,
        nodeHeight,
        direction,
      );
      setDags(builtDags);
      setLoadingDags(false);
    },
    onError: () => {
      setLoadingDags(false);
    },
  });

  return {
    error,
    dag: dags,
    loading: loading || loadingDags,
  };
};
