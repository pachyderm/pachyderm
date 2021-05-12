import {gql} from 'graphql-tag';

export const GET_ACCOUNT_QUERY = gql`
  query getAccount {
    account {
      id
      email
      name
    }
  }
`;
