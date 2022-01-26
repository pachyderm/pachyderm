import {gql} from '@apollo/client';

export const GET_PIPELINES_QUERY = gql`
  query pipelines($args: PipelinesQueryArgs!) {
    pipelines(args: $args) {
      id
      name
      state
      type
      description
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
