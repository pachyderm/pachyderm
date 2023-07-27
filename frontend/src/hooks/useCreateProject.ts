import {CreateProjectArgs, Project} from '@graphqlTypes';

import {useCreateProjectMutation} from '@dash-frontend/generated/hooks';
import {GET_PROJECTS_QUERY} from '@dash-frontend/queries/GetProjectsQuery';

const useCreateProject = (onCompleted?: () => void) => {
  const [createProjectMutation, mutationResult] = useCreateProjectMutation({
    onCompleted,
    update(cache, {data}) {
      if (!data) return;
      cache.modify({
        fields: {
          projects(existingProjects: Project[]) {
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
