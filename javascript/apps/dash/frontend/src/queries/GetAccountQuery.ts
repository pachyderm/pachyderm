import {gql} from '@apollo/client';

export const GET_ACCOUNT_QUERY = gql`
  query getAccount {
    account {
      id
      email
      name
    }
  }
`;
