import {CreateProjectArgs} from '@graphqlTypes';

import {useCreateProjectMutation} from '@dash-frontend/generated/hooks';

const useCreateProject = (onCompleted?: () => void) => {
  const [createProjectMutation, mutationResult] = useCreateProjectMutation({
    onCompleted,
  });

  return {
    createProject: (args: CreateProjectArgs) =>
      createProjectMutation({variables: {args}}),
    ...mutationResult,
  };
};

export default useCreateProject;
