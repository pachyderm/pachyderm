import {useQuery} from '@apollo/client';

import {Dag, QueryDagArgs} from '@graphqlTypes';
import {GET_DAG_QUERY} from 'queries/GetDagQuery';

type DagQueryResponse = {
  dag: Dag;
};

export const useDAGData = (projectId = '') => {
  const {data, error, loading} = useQuery<DagQueryResponse, QueryDagArgs>(
    GET_DAG_QUERY,
    {variables: {args: {projectId}}},
  );

  return {
    error,
    dag: data?.dag,
    loading,
  };
};
