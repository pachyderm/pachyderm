import {gql} from '@apollo/client';

export const CREATE_PIPELINE_MUTATION = gql`
  mutation createPipeline($args: CreatePipelineArgs!) {
    createPipeline(args: $args) {
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

export default CREATE_PIPELINE_MUTATION;
