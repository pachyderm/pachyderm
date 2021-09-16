import {gql} from '@apollo/client';

export const CREATE_REPO_MUTATION = gql`
  mutation createRepo($args: CreateRepoArgs!) {
    createRepo(args: $args) {
      createdAt
      description
      id
      name
      sizeDisplay
    }
  }
`;

export default CREATE_REPO_MUTATION;
