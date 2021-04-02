import {useProjectQuery} from '@dash-frontend/generated/hooks';

interface UseProjectArgs {
  id: string;
}

const useProject = ({id}: UseProjectArgs) => {
  const {data, error, loading} = useProjectQuery({
    variables: {id},
  });

  return {
    project: data?.project,
    error,
    loading,
  };
};

export default useProject;
