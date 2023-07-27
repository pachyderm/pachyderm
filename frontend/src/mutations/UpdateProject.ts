import {gql} from '@apollo/client';

export const UPDATE_PROJECT_MUTATION = gql`
  mutation updateProject($args: UpdateProjectArgs!) {
    updateProject(args: $args) {
      id
      description
    }
  }
`;

export default UPDATE_PROJECT_MUTATION;
