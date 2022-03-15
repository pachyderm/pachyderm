import {gql} from '@apollo/client';

export const CommitFragment = gql`
  fragment CommitFragment on Commit {
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
`;
