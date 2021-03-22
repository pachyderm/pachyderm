import {gql} from '@apollo/client';

export const GET_DAG_QUERY = gql`
  query getDag($args: DagQueryArgs!) {
    dag(args: $args) {
      nodes {
        name
        type
        access
        state
      }
      links {
        source
        target
        state
      }
    }
  }
`;
