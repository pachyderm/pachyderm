import {gql} from '@apollo/client';

export const GET_DAG_QUERY = gql`
  query getDag($args: DagQueryArgs!) {
    dag(args: $args) {
      name
      state
      access
      parents
      type
      jobState
      createdAt
    }
  }
`;
