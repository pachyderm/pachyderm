import {gql} from '@apollo/client';

export const GET_DAG_QUERY = gql`
  query getDag($args: DagQueryArgs!) {
    dag(args: $args) {
      id
      name
      state
      nodeState
      access
      parents
      type
      jobState
      jobNodeState
      createdAt
    }
  }
`;
