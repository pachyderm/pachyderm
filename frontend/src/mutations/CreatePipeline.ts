import {gql} from '@apollo/client';

import {PipelineFragment} from '@dash-frontend/fragments/Pipeline';

export const CREATE_PIPELINE_MUTATION = gql`
  mutation createPipeline($args: CreatePipelineArgs!) {
    createPipeline(args: $args) {
      ...PipelineFragment
    }
  }
  ${PipelineFragment}
`;

export default CREATE_PIPELINE_MUTATION;
