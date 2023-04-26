import {gql} from '@apollo/client';

export const FIND_COMMITS_QUERY = gql`
  query findCommits($args: FindCommitsQueryArgs!) {
    findCommits(args: $args) {
      commits {
        id
        started
        commitAction
      }
      cursor
      hasNextPage
    }
  }
`;
