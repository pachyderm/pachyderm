import {gql} from '@apollo/client';

import {DatumFragment} from '@dash-frontend/fragments/Datum';

export const DATUM_SEARCH_QUERY = gql`
  query datumSearch($args: DatumQueryArgs!) {
    datumSearch(args: $args) {
      ...Datum
    }
  }
  ${DatumFragment}
`;
