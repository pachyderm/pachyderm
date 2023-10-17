import {gql} from '@apollo/client';

import {PipelineFragment} from '@dash-frontend/fragments/Pipeline';

export const GET_PIPELINES_QUERY = gql`
  query pipelines($args: PipelinesQueryArgs!) {
    pipelines(args: $args) {
      ...PipelineFragment
    }
  }
  ${PipelineFragment}
`;
