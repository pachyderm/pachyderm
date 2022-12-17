import {QueryFunctionOptions} from '@apollo/client';
import {DatumQueryArgs} from '@graphqlTypes';

import {useDatumSearchQuery} from '@dash-frontend/generated/hooks';

const useDatumSearch = (args: DatumQueryArgs, opts?: QueryFunctionOptions) => {
  const {data, error, loading} = useDatumSearchQuery({
    variables: {args},
    skip: opts?.skip,
  });

  return {
    error,
    datum: data?.datumSearch,
    loading,
  };
};

export default useDatumSearch;
