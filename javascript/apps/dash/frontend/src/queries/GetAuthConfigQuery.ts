import {gql} from '@apollo/client';

export const GET_AUTH_CONFIG_QUERY = gql`
  query authConfig {
    authConfig {
      authUrl
      clientId
      pachdClientId
    }
  }
`;
