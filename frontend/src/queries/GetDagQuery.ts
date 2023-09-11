import {gql} from '@apollo/client';

export const GET_DAG_QUERY = gql`
  query getDag($args: DagQueryArgs!) {
    dag(args: $args) {
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
