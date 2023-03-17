import {useProjectQuery} from '@dash-frontend/generated/hooks';

interface UseProjectArgs {
  id: string;
}

const useProject = ({id}: UseProjectArgs) => {
  const projectQuery = useProjectQuery({
    variables: {id},
  });

  return {
    ...projectQuery,
    project: projectQuery?.data?.project,
  };
};

export default useProject;
