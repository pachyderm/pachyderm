import {gql} from '@apollo/client';

import {DatumFragment} from '@dash-frontend/fragments/Datum';

export const DATUM_QUERY = gql`
  query datum($args: DatumQueryArgs!) {
    datum(args: $args) {
      ...Datum
    }
  }
  ${DatumFragment}
`;
