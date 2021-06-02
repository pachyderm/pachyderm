import {gql} from '@apollo/client';

export const PipelineJobOverviewFragment = gql`
  fragment PipelineJobOverview on PipelineJob {
    id
    state
    createdAt
    finishedAt
  }
`;
