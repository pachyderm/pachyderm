import {gql} from '@apollo/client';

export const GET_DAGS_QUERY = gql`
  query getDags($args: DagQueryArgs!) {
    dags(args: $args) {
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
