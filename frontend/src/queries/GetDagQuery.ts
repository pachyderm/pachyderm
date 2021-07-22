import {gql} from '@apollo/client';

export const GET_DAG_QUERY = gql`
  query getDag($args: DagQueryArgs!) {
    dag(args: $args) {
      nodes {
        id
        name
        type
        access
        state
        x
        y
      }
      links {
        id
        source
        target
        sourceState
        targetState
        state
        bendPoints {
          x
          y
        }
        startPoint {
          x
          y
        }
        endPoint {
          x
          y
        }
      }
      id
    }
  }
`;
