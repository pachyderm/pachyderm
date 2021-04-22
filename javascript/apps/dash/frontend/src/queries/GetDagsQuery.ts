import {gql} from '@apollo/client';

export const GET_DAGS_QUERY = gql`
  subscription getDags($args: DagQueryArgs!) {
    dags(args: $args) {
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
      priorityPipelineState
    }
  }
`;
