import {CreateProjectArgs} from '@graphqlTypes';

import {useCreateProjectMutation} from '@dash-frontend/generated/hooks';

const useCreateProject = (onCompleted?: () => void) => {
  const [createProjectMutation, mutationResult] = useCreateProjectMutation({
    onCompleted,
    update(cache, {data}) {
      if (!data) return;
      cache.modify({
        fields: {
          projects(existingProjects) {
            return [...existingProjects, data.createProject];
          },
        },
      });
    },
  });

  return {
    createProject: (args: CreateProjectArgs) =>
      createProjectMutation({variables: {args}}),
    ...mutationResult,
  };
};

export default useCreateProject;
