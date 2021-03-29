import {gql} from '@apollo/client';

export const GET_LOGGED_IN_QUERY = gql`
  query loggedIn {
    loggedIn @client
  }
`;
