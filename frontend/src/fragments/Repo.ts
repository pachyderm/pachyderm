import {gql} from '@apollo/client';

export const RepoFragment = gql`
  fragment RepoFragment on Repo {
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
  }
`;
