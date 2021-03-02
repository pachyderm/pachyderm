import gql from 'graphql-tag';

export const GET_PROJECTS_QUERY = gql`
  query projects {
    projects {
      id
      name
      description
      createdAt
      status
    }
  }
`;
