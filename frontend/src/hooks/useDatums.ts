import {DatumsQueryArgs} from '@graphqlTypes';

import {useDatumsQuery} from '@dash-frontend/generated/hooks';

const useDatums = (args: DatumsQueryArgs) => {
  const {data, error, loading, refetch} = useDatumsQuery({
    variables: {args},
  });

  return {
    error,
    datums: data?.datums.items || [],
    cursor: data?.datums.cursor,
    hasNextPage: data?.datums.hasNextPage,
    loading,
    refetch,
  };
};

export default useDatums;
