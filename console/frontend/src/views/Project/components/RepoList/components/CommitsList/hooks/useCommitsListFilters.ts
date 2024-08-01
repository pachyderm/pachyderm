import {useEffect} from 'react';
import {useForm} from 'react-hook-form';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';

export const commitsFilters = [
  {
    label: 'Sort By',
    name: 'sortBy',
    type: 'radio',
    options: [
      {
        name: 'Created: Newest',
        value: 'Created: Newest',
      },
      {
        name: 'Created: Oldest',
        value: 'Created: Oldest',
      },
    ],
  },
];

type FormValues = {
  sortBy: string;
};

const useCommitsListFilters = () => {
  const {getNewSearchParamsAndGo, searchParams} = useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
    },
  });

  const {watch} = formCtx;
  const sortFilter = watch('sortBy');
  const reverseOrder = sortFilter === 'Created: Oldest';

  useEffect(() => {
    const {selectedRepos, branchId} = searchParams;
    getNewSearchParamsAndGo({
      sortBy: sortFilter,
      selectedRepos,
      branchId,
    });
  }, [sortFilter, getNewSearchParamsAndGo, searchParams]);

  const staticFilterKeys = [sortFilter];

  return {
    formCtx,
    commitsFilters,
    staticFilterKeys,
    reverseOrder,
  };
};

export default useCommitsListFilters;
