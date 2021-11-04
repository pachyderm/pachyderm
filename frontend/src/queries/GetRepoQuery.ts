import {gql} from '@apollo/client';

export const GET_REPO_QUERY = gql`
  query repo($args: RepoQueryArgs!) {
    repo(args: $args) {
      branches {
        name
      }
      commits {
        repoName
        branch {
          name
        }
        originKind
        description
        hasLinkedJob
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
