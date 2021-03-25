import {useGetDagsQuery} from '@dash-frontend/generated/hooks';

export const useProjectDagsData = (projectId = '') => {
  const {data, error, loading} = useGetDagsQuery({
    variables: {args: {projectId}},
  });

  return {
    error,
    dags: data?.dags,
    loading,
  };
};
