import {gql} from '@apollo/client';

export const DatumFragment = gql`
  fragment Datum on Datum {
    id
    jobId
    requestedJobId
    state
    downloadTimestamp {
      seconds
      nanos
    }
    uploadTimestamp {
      seconds
      nanos
    }
    processTimestamp {
      seconds
      nanos
    }
    downloadBytes
    uploadBytes
  }
`;
