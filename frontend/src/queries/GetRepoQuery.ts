import {gql} from '@apollo/client';

import {RepoFragment} from '@dash-frontend/fragments/Repo';

export const GET_REPO_QUERY = gql`
  query repo($args: RepoQueryArgs!) {
    repo(args: $args) {
      ...RepoFragment
    }
  }
  ${RepoFragment}
`;
