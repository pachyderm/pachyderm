import {gql} from '@apollo/client';

import {CommitFragment} from '@dash-frontend/fragments/Commit';
import {RepoFragment} from '@dash-frontend/fragments/Repo';

export const GET_REPOS_WITH_COMMIT_QUERY = gql`
  query reposWithCommit($args: ReposQueryArgs!) {
    repos(args: $args) {
      ...RepoFragment
      lastCommit {
        ...CommitFragment
      }
    }
  }
  ${RepoFragment}
  ${CommitFragment}
`;
