import {gql} from '@apollo/client';

export const DiffFragment = gql`
  fragment DiffFragment on Diff {
    size
    sizeDisplay
    filesUpdated
    filesAdded
    filesDeleted
  }
`;
