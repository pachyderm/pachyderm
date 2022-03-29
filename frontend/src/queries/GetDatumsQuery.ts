import {gql} from '@apollo/client';

import {DatumFragment} from '@dash-frontend/fragments/Datum';

export const DATUMS_QUERY = gql`
  query datums($args: DatumsQueryArgs!) {
    datums(args: $args) {
      ...Datum
    }
  }
  ${DatumFragment}
`;
