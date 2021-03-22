import {gql} from '@apollo/client';

export const EXCHANGE_CODE_MUTATION = gql`
  mutation exchangeCode($code: String!) {
    exchangeCode(code: $code) {
      pachToken
      idToken
    }
  }
`;

export default EXCHANGE_CODE_MUTATION;
