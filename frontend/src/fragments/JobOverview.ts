import {gql} from '@apollo/client';

export const JobOverviewFragment = gql`
  fragment JobOverview on Job {
    id
    state
    nodeState
    createdAt
    startedAt
    finishedAt
    restarts
    pipelineName
    reason
    dataProcessed
    dataSkipped
    dataFailed
    dataTotal
    dataRecovered
    downloadBytesDisplay
    uploadBytesDisplay
    outputCommit
  }
`;
