import {gql} from '@apollo/client';

export const GET_PERMISSIONS_QUERY = gql`
  query getPermissions($args: GetPermissionsArgs!) {
    getPermissions(args: $args) {
      rolesList
    }
  }
`;
