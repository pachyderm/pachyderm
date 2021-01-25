import gql from 'graphql-tag';

export const GET_DAG_QUERY = gql`
  query getDag($args: DagQueryArgs!) {
    dag(args: $args) {
      nodes {
        name
        type
        error
        access
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
