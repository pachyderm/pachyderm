import {gql} from '@apollo/client';

export const CREATE_PIPELINE_V2_MUTATION = gql`
  mutation createPipelineV2($args: CreatePipelineV2Args!) {
    createPipelineV2(args: $args) {
      effectiveCreatePipelineRequestJson
    }
  }
`;

export default CREATE_PIPELINE_V2_MUTATION;
