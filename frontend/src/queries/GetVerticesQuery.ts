import {gql} from '@apollo/client';

export const GET_VERTICES_QUERY = gql`
  query getVertices($args: VerticesQueryArgs!) {
    vertices(args: $args) {
      id
      project
      name
      parents {
        id
        project
        name
      }
      state
      nodeState
      access
      type
      jobState
      jobNodeState
      createdAt
    }
  }
`;
