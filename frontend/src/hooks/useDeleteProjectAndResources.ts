import {DeleteProjectAndResourcesArgs} from '@graphqlTypes';

import {useDeleteProjectAndResourcesMutation} from '@dash-frontend/generated/hooks';
import {GET_PIPELINES_QUERY} from '@dash-frontend/queries/GetPipelinesQuery';
import {GET_PROJECTS_QUERY} from '@dash-frontend/queries/GetProjectsQuery';
import {GET_REPOS_QUERY} from '@dash-frontend/queries/GetReposQuery';

export const useDeleteProjectAndResources = (
  onCompleted?: () => void,
  projectId?: string,
) => {
  const [deleteProjectMutation, mutationResult] =
    useDeleteProjectAndResourcesMutation({
      onCompleted,
      refetchQueries: () => [
        {query: GET_PROJECTS_QUERY},
        {
          query: GET_PIPELINES_QUERY,
          variables: {args: {projectIds: [projectId]}},
        },
        {
          query: GET_REPOS_QUERY,
          variables: {args: {projectId}},
        },
      ],
      awaitRefetchQueries: true,
    });

  return {
    deleteProject: (args: DeleteProjectAndResourcesArgs) =>
      deleteProjectMutation({variables: {args}}),
    ...mutationResult,
  };
};
