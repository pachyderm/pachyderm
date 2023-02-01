import {gql} from '@apollo/client';

export const GET_REPO_QUERY = gql`
  query repo($args: RepoQueryArgs!) {
    repo(args: $args) {
      branches {
        name
      }
      createdAt
      description
      id
      linkedPipeline {
        id
        name
      }
      name
      sizeDisplay
      projectId
    }
  }
`;
