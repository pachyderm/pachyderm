import {Datum, DatumsQueryArgs, DatumsQuery} from '@graphqlTypes';

import {useDatumsQuery} from '@dash-frontend/generated/hooks';

const useDatums = (args: DatumsQueryArgs) => {
  const {data, error, loading, refetch} = useDatumsQuery({
    variables: {args},
  });

  // This is to assure that the types coming back are all of type Datum
  const datumTypeGuard = (items: DatumsQuery['datums']['items']): Datum[] => {
    const result = items.every((item) => {
      return item?.__typename === 'Datum';
    });
    if (!result) {
      console.error('Datums Type Error');
    }
    return items as Datum[];
  };

  return {
    error,
    datums: datumTypeGuard(data?.datums.items || []),
    cursor: data?.datums.cursor,
    hasNextPage: data?.datums.hasNextPage,
    loading,
    refetch,
  };
};

export default useDatums;
