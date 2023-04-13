import {gql} from '@apollo/client';

import {CommitFragment} from '@dash-frontend/fragments/Commit';

export const FIND_COMMITS_QUERY = gql`
  query findCommits($args: FindCommitsQueryArgs!) {
    findCommits(args: $args) {
      commits {
        ...CommitFragment
      }
      cursor
      hasNextPage
    }
  }
  ${CommitFragment}
`;
