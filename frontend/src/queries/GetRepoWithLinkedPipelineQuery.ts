import {gql} from '@apollo/client';

import {RepoWithLinkedPipelineFragment} from '@dash-frontend/fragments/RepoWithLinkedPipeline';

export const GET_REPO_WITH_LINKED_PIPELINE_QUERY = gql`
  query repoWithLinkedPipeline($args: RepoQueryArgs!) {
    repo(args: $args) {
      ...RepoWithLinkedPipelineFragment
    }
  }
  ${RepoWithLinkedPipelineFragment}
`;
