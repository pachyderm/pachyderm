import {gql} from '@apollo/client';

export const GET_DAGS_QUERY = gql`
  subscription getDags($args: DagQueryArgs!) {
    dags(args: $args) {
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
