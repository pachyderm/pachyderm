import {gql} from '@apollo/client';

export const CREATE_PROJECT_MUTATION = gql`
  mutation createProject($args: CreateProjectArgs!) {
    createProject(args: $args) {
      id
      description
      status
      createdAt {
        seconds
        nanos
      }
    }
  }
`;

export default CREATE_PROJECT_MUTATION;
