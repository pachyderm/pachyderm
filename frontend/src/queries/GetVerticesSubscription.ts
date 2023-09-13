import {gql} from '@apollo/client';

export const GET_VERTICES_SUBSCRIPTION = gql`
  subscription vertices($args: VerticesQueryArgs!) {
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
