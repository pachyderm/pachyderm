import {gql} from '@apollo/client';

export const GET_AUTHORIZE = gql`
  query getAuthorize($args: GetAuthorizeArgs!) {
    getAuthorize(args: $args) {
      satisfiedList
      missingList
      authorized
      principal
    }
  }
`;
