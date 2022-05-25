import {gql} from '@apollo/client';

export const GET_ENTERPRISE_INFO_QUERY = gql`
  query getEnterpriseInfo {
    enterpriseInfo {
      state
      expiration
    }
  }
`;
