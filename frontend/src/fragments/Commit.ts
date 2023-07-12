import {gql} from '@apollo/client';

export const CommitFragment = gql`
  fragment CommitFragment on Commit {
    repoName
    branch {
      name
      repo {
        name
        type
      }
    }
    description
    originKind
    id
    started
    finished
    sizeBytes
    sizeDisplay
  }
`;
