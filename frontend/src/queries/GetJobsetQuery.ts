import {gql} from '@apollo/client';

import {JobSetFragment} from '@dash-frontend/fragments/JobSet';

export const JOB_SET_QUERY = gql`
  query jobSet($args: JobSetQueryArgs!) {
    jobSet(args: $args) {
      ...JobSetFields
    }
  }
  ${JobSetFragment}
`;
