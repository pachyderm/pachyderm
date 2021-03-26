import {gql} from '@apollo/client';

export const GET_PROJECT_DETAILS_QUERY = gql`
  query projectDetails($args: ProjectDetailsQueryArgs!) {
    projectDetails(args: $args) {
      sizeGBytes
      repoCount
      pipelineCount
      jobs {
        id
        state
        createdAt
      }
    }
  }
`;
