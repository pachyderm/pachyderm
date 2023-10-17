import {gql} from '@apollo/client';

export const PipelineFragment = gql`
  fragment PipelineFragment on Pipeline {
    id
    name
    description
    version
    createdAt
    state
    nodeState
    stopped
    recentError
    lastJobState
    lastJobNodeState
    type
    datumTimeoutS
    datumTries
    jobTimeoutS
    outputBranch
    s3OutputRepo
    egress
    userSpecJson
    effectiveSpecJson
    parallelismSpec
    reason
  }
`;
