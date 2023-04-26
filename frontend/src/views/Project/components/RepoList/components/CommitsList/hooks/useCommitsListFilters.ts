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
  const {searchParams, updateSearchParamsAndGo, getNewSearchParamsAndGo} =
    useUrlQueryState();
  const formCtx = useForm<FormValues>({
    mode: 'onChange',
    defaultValues: {
      sortBy: 'Created: Newest',
    },
  });

  const {watch, reset} = formCtx;
  const sortFilter = watch('sortBy');
  const reverseOrder = sortFilter === 'Created: Oldest';

  useEffect(() => {
    const {selectedRepos} = searchParams;
    reset();
    getNewSearchParamsAndGo({
      selectedRepos,
    });
    // We want to clear viewstate and form state on fresh renders
    // but NOT on changes to searchParams
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [getNewSearchParamsAndGo, reset, updateSearchParamsAndGo]);

  useEffect(() => {
    updateSearchParamsAndGo({
      sortBy: sortFilter,
    });
  }, [sortFilter, updateSearchParamsAndGo]);

  const staticFilterKeys = [sortFilter];

  return {
    formCtx,
    commitsFilters,
    staticFilterKeys,
    reverseOrder,
  };
};

export default useCommitsListFilters;
