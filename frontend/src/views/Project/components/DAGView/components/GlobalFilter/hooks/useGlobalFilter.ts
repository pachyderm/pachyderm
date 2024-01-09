import {useCallback, useState, useRef, useEffect} from 'react';
import {useForm} from 'react-hook-form';

import {useJobSetLazy} from '@dash-frontend/hooks/useJobSet';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {useOutsideClick} from '@pachyderm/components';

type GlobalIdFilterFormValues = {
  globalId: string;
};

const useGlobalFilter = () => {
  const {searchParams, updateSearchParamsAndGo} = useUrlQueryState();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const globalIdFilter = searchParams.globalIdFilter;
  const formCtx = useForm<GlobalIdFilterFormValues>({
    mode: 'onChange',
    defaultValues: {globalId: globalIdFilter},
  });
  const {
    watch,
    setError,
    reset,
    getValues,
    formState,
    getFieldState,
    setValue,
  } = formCtx;
  const globalIdInput = watch('globalId');

  const {
    refetch: getJobSet,
    jobSet,
    loading,
    ...rest
  } = useJobSetLazy(globalIdInput);

  const clearFilter = useCallback(() => {
    updateSearchParamsAndGo({
      globalIdFilter: undefined,
    });
    getJobSet();
    setDropdownOpen(false);
    reset({globalId: ''});
  }, [getJobSet, reset, updateSearchParamsAndGo]);

  const handleSubmit = useCallback(() => {
    const formGlobalId = getValues('globalId');
    const fieldState = getFieldState('globalId');

    if (
      (formGlobalId === '' && globalIdFilter) ||
      formGlobalId === globalIdFilter
    ) {
      clearFilter();
    } else if (!fieldState.error) {
      getJobSet();
    }
  }, [clearFilter, getFieldState, getJobSet, getValues, globalIdFilter]);

  useEffect(() => {
    if (globalIdFilter) {
      setValue('globalId', globalIdFilter);
    }
  }, [globalIdFilter, setValue]);

  // apply viewstate if the jobset id returned jobs
  useEffect(() => {
    if (jobSet?.length && jobSet?.[0]?.job?.id) {
      updateSearchParamsAndGo({
        globalIdFilter: jobSet[0].job.id,
      });
      setDropdownOpen(false);
    }
  }, [jobSet, loading, setError, updateSearchParamsAndGo]);

  // set error if we queried for the jobset but no jobs were returned
  useEffect(() => {
    const inCorrectQueryState =
      rest.isFetched === true &&
      rest.isRefetching === false &&
      rest.isSuccess === true &&
      loading === false;
    if (!formState.errors.globalId && inCorrectQueryState && !jobSet?.length) {
      setError('globalId', {
        type: 'jobsetError',
        message: 'This Global ID does not exist',
      });
    }
  }, [
    formState.errors.globalId,
    globalIdInput,
    jobSet?.length,
    loading,
    rest.isFetched,
    rest.isRefetching,
    rest.isSuccess,
    setError,
  ]);

  const handleOutsideClick = useCallback(() => {
    if (dropdownOpen) {
      setDropdownOpen(false);
      reset({globalId: globalIdFilter});
    }
  }, [dropdownOpen, globalIdFilter, reset]);

  useOutsideClick(containerRef, handleOutsideClick);

  return {
    containerRef,
    formCtx,
    dropdownOpen,
    setDropdownOpen,
    loading,
    globalIdFilter,
    globalIdInput,
    handleSubmit,
  };
};

export default useGlobalFilter;
