import {gql} from '@apollo/client';

export const GET_PROJECT_QUERY = gql`
  query project($id: ID!) {
    project(id: $id) {
      id
      name
      description
      createdAt
      status
    }
  }
`;
