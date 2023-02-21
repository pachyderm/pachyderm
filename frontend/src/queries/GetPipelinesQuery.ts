import {gql} from '@apollo/client';

export const GET_PIPELINES_QUERY = gql`
  query pipelines($args: PipelinesQueryArgs!) {
    pipelines(args: $args) {
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
