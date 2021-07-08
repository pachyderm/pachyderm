import {gql} from '@apollo/client';

import {JobSetFragment} from '@dash-frontend/fragments/JobSet';

export const JOB_SETS_QUERY = gql`
  query jobSets($args: JobSetsQueryArgs!) {
    jobSets(args: $args) {
      ...JobSetFields
    }
  }
  ${JobSetFragment}
`;
