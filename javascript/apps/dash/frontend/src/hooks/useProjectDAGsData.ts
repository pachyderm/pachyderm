import {useQuery} from '@apollo/client';

import {Dag, QueryDagArgs} from '@graphqlTypes';
import {GET_DAGS_QUERY} from 'queries/GetDagsQuery';

type DagsQueryResponse = {
  dags: [Dag];
};

export const useProjectDagsData = (projectId = '') => {
  const {data, error, loading} = useQuery<DagsQueryResponse, QueryDagArgs>(
    GET_DAGS_QUERY,
    {variables: {args: {projectId}}},
  );

  return {
    error,
    dags: data?.dags,
    loading,
  };
};
