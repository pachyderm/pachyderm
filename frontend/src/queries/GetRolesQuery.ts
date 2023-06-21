import {gql} from '@apollo/client';

export const GET_ROLES_QUERY = gql`
  query getRoles($args: GetRolesArgs!) {
    getRoles(args: $args) {
      roleBindings {
        principal
        roles
      }
    }
  }
`;
