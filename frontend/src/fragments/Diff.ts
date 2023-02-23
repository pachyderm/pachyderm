import {gql} from '@apollo/client';

export const DiffFragment = gql`
  fragment DiffFragment on Diff {
    size
    sizeDisplay
    filesUpdated {
      count
      sizeDelta
    }
    filesAdded {
      count
      sizeDelta
    }
    filesDeleted {
      count
      sizeDelta
    }
  }
`;
