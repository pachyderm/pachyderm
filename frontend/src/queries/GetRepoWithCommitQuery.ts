import {gql} from '@apollo/client';

import {CommitFragment} from '@dash-frontend/fragments/Commit';
import {RepoFragment} from '@dash-frontend/fragments/Repo';

export const GET_REPO_WITH_COMMIT_QUERY = gql`
  query repoWithCommit($args: RepoQueryArgs!) {
    repo(args: $args) {
      ...RepoFragment
      lastCommit {
        ...CommitFragment
      }
    }
  }
  ${RepoFragment}
  ${CommitFragment}
`;
