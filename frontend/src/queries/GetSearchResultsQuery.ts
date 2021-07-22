import {gql} from '@apollo/client';

export const GET_SEARCH_RESULTS_QUERY = gql`
  query searchResults($args: SearchResultQueryArgs!) {
    searchResults(args: $args) {
      pipelines {
        name
        id
      }
      repos {
        name
        id
      }
      jobSet {
        id
      }
    }
  }
`;
