import {useQuery} from '@apollo/client';

import {GET_ACCOUNT_QUERY} from '@dash-frontend/queries/GetAccountQuery';
import {Account} from '@graphqlTypes';

type AccountQueryResponse = {
  account: Account;
};

interface useAccountArgs {
  skip?: boolean;
}

const useAccount = ({skip = false}: useAccountArgs = {}) => {
  const {data, error, loading} = useQuery<AccountQueryResponse>(
    GET_ACCOUNT_QUERY,
    {skip},
  );

  return {
    error,
    account: data?.account,
    displayName: data?.account.name || data?.account.email,
    loading,
  };
};

export default useAccount;
