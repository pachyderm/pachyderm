import {gql} from '@apollo/client';

export const LogFragment = gql`
  fragment LogFields on Log {
    timestamp {
      seconds
      nanos
    }
    user
    message
  }
`;
