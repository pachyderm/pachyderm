import {useGetDagQuery} from '@dash-frontend/generated/hooks';

export const useDAGData = (projectId = '') => {
  const {data, error, loading} = useGetDagQuery({
    variables: {args: {projectId}},
  });

  return {
    error,
    dag: data?.dag,
    loading,
  };
};
