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
  direction = DagDirection.DOWN,
}: GetDagQueryProps) => {
  const [loadingDags, setLoadingDags] = useState(true);
  const [dagError, setDagError] = useState<string>();
  const [dags, setDags] = useState<Dag[] | undefined>();
  const {error, loading} = useGetDagQuery({
    variables: {args: {projectId}},
    onCompleted: async (data) => {
      const builtDags = await buildDags(
        data.dag,
        nodeWidth,
        nodeHeight,
        direction,
        setDagError,
      );
      setDags(builtDags);
      setLoadingDags(false);
    },
    onError: () => {
      setLoadingDags(false);
    },
  });

  return {
    error: error || dagError,
    dag: dags,
    loading: loading || loadingDags,
  };
};
