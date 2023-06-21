import {gql} from '@apollo/client';

export const MODIFY_ROLES_MUTATION = gql`
  mutation modifyRoles($args: ModifyRolesArgs!) {
    modifyRoles(args: $args)
  }
`;

export default MODIFY_ROLES_MUTATION;
