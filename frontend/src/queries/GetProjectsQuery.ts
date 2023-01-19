import {gql} from '@apollo/client';

export const GET_PROJECTS_QUERY = gql`
  query projects {
    projects {
      id
      description
      status
    }
  }
`;
