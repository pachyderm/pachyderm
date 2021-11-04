import {gql} from '@apollo/client';

export const CREATE_BRANCH_MUTATION = gql`
  mutation createBranch($args: CreateBranchArgs!) {
    createBranch(args: $args) {
      name
      repo {
        name
      }
    }
  }
`;

export default CREATE_BRANCH_MUTATION;
