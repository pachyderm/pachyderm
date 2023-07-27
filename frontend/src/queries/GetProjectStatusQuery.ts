import {gql} from '@apollo/client';

export const GET_PROJECT_STATUS_QUERY = gql`
  query projectStatus($id: ID!) {
    projectStatus(id: $id) {
      id
      status
    }
  }
`;
