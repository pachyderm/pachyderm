import {gql} from '@apollo/client';

export const DatumFragment = gql`
  fragment Datum on Datum {
    id
    state
    downloadBytes
    uploadTime
    processTime
    downloadTime
  }
`;
