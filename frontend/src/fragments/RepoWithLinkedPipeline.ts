import {gql} from '@apollo/client';

export const RepoWithLinkedPipelineFragment = gql`
  fragment RepoWithLinkedPipelineFragment on Repo {
    branches {
      name
    }
    createdAt
    description
    id
    name
    sizeDisplay
    sizeBytes
    access
    projectId
    linkedPipeline {
      id
      name
    }
    authInfo {
      rolesList
    }
  }
`;
