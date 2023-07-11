import {gql} from '@apollo/client';

import {RepoWithLinkedPipelineFragment} from '@dash-frontend/fragments/RepoWithLinkedPipeline';

export const GET_REPOS_WITH_LINKED_PIPELINE_QUERY = gql`
  query reposWithLinkedPipeline($args: ReposQueryArgs!) {
    repos(args: $args) {
      ...RepoWithLinkedPipelineFragment
    }
  }
  ${RepoWithLinkedPipelineFragment}
`;
