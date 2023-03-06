import {gql} from '@apollo/client';

import {CommitFragment} from '@dash-frontend/fragments/Commit';

export const GET_COMMITS_QUERY = gql`
  query getCommits($args: CommitsQueryArgs!) {
    commits(args: $args) {
      items {
        ...CommitFragment
      }
      cursor {
        seconds
        nanos
      }
      hasNextPage
    }
  }
  ${CommitFragment}
`;
