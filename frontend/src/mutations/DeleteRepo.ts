import {gql} from '@apollo/client';

export const DELETE_REPO_MUTATION = gql`
  mutation deleteRepo($args: DeleteRepoArgs!) {
    deleteRepo(args: $args)
  }
`;

export default DELETE_REPO_MUTATION;
