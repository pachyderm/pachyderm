import {QueryFunctionOptions} from '@apollo/client';
import {DatumQueryArgs} from '@graphqlTypes';

import {useDatumQuery} from '@dash-frontend/generated/hooks';

const useDatum = (args: DatumQueryArgs, opts?: QueryFunctionOptions) => {
  const {data, error, loading} = useDatumQuery({
    variables: {args},
    skip: opts?.skip,
  });

  return {
    error,
    datum: data?.datum,
    loading,
  };
};

export default useDatum;
