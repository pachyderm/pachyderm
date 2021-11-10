import {gql} from '@apollo/client';

export const GET_COMMITS_QUERY = gql`
  query getCommits($args: CommitsQueryArgs!) {
    commits(args: $args) {
      repoName
      branch {
        name
      }
      description
      originKind
      id
      started
      finished
      sizeBytes
      sizeDisplay
      hasLinkedJob
    }
  }
`;
