import {gql} from '@apollo/client';

export const BranchFragment = gql`
  fragment BranchFragment on Branch {
    name
    repo {
      name
      type
    }
  }
`;
