import {gql} from '@apollo/client';

export const JobOverviewFragment = gql`
  fragment JobOverview on Job {
    id
    state
    createdAt
    finishedAt
    pipelineName
  }
`;
