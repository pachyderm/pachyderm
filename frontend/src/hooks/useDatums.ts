import {Datum, DatumsQueryArgs, DatumsQuery} from '@graphqlTypes';

import {DATUMS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useDatumsQuery} from '@dash-frontend/generated/hooks';

const useDatums = (args: DatumsQueryArgs) => {
  const {data, error, loading} = useDatumsQuery({
    variables: {args},
    pollInterval: DATUMS_POLL_INTERVAL_MS,
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
  };
};

export default useDatums;
