import {gql} from '@apollo/client';

export const GET_DAGS_QUERY = gql`
  subscription getDags($args: DagQueryArgs!) {
    dags(args: $args) {
      id
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
