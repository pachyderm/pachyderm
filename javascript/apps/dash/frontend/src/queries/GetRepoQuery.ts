import {gql} from '@apollo/client';

export const GET_REPO_QUERY = gql`
  query repo($args: RepoQueryArgs!) {
    repo(args: $args) {
      branches {
        id
        name
      }
      commits {
        repoName
        branch {
          id
          name
        }
        description
        id
        started
        finished
        sizeDisplay
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
    }
  }
`;
