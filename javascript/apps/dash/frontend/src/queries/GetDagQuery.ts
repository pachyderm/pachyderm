import gql from 'graphql-tag';

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
        error
        active
      }
    }
  }
`;
