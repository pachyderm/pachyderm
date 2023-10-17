import {gql} from '@apollo/client';

import {PipelineFragment} from '@dash-frontend/fragments/Pipeline';

export const GET_PIPELINE_QUERY = gql`
  query pipeline($args: PipelineQueryArgs!) {
    pipeline(args: $args) {
      ...PipelineFragment
    }
  }
  ${PipelineFragment}
`;
