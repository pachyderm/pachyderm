import {gql} from '@apollo/client';

export const GET_ADMIN_INFO_QUERY = gql`
  query getAdminInfo {
    adminInfo {
      clusterId
    }
  }
`;
