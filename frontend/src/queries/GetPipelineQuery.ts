import {gql} from '@apollo/client';

export const GET_PIPELINE_QUERY = gql`
  query pipeline($args: PipelineQueryArgs!) {
    pipeline(args: $args) {
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
      jsonSpec
      reason
    }
  }
`;
