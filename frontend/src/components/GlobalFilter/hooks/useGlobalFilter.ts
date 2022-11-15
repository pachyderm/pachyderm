import {useCallback, useState, useRef, useEffect} from 'react';
import {useForm} from 'react-hook-form';

import {useJobSetLazyQuery} from '@dash-frontend/generated/hooks';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useOutsideClick} from '@pachyderm/components';

type GlobalIdFilterFormValues = {
  globalId: string;
};

const useGlobalFilter = () => {
  const {viewState, updateViewState} = useUrlQueryState();
  const {projectId} = useUrlState();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const [getJobSet, {data, loading}] = useJobSetLazyQuery();

  const globalIdFilter = viewState.globalIdFilter;

  const formCtx = useForm<GlobalIdFilterFormValues>({
    mode: 'onChange',
    defaultValues: {globalId: globalIdFilter},
  });

  const {watch, setError, reset, getValues, formState, getFieldState} = formCtx;

  const globalIdInput = watch('globalId');

  const clearFilter = useCallback(() => {
    updateViewState({
      globalIdFilter: undefined,
    });
    getJobSet({variables: {args: {projectId, id: ''}}});
    setDropdownOpen(false);
    reset({globalId: ''});
  }, [getJobSet, projectId, reset, updateViewState]);

  const handleSubmit = useCallback(() => {
    const formGlobalId = getValues('globalId');
    const fieldState = getFieldState('globalId');

    if (
      (formGlobalId === '' && globalIdFilter) ||
      formGlobalId === globalIdFilter
    ) {
      clearFilter();
    } else if (!fieldState.error) {
      getJobSet({variables: {args: {projectId, id: getValues('globalId')}}});
    }
  }, [
    clearFilter,
    getFieldState,
    getJobSet,
    getValues,
    globalIdFilter,
    projectId,
  ]);

  // apply viewstate if the jobset id returned jobs
  useEffect(() => {
    if (data?.jobSet.jobs.length) {
      updateViewState({
        globalIdFilter: data?.jobSet.id,
      });
      setDropdownOpen(false);
    }
  }, [data, loading, setError, updateViewState]);

  // set error if we queried for the jobset but no jobs were returned
  useEffect(() => {
    if (
      !formState.errors.globalId &&
      data?.jobSet.id &&
      data?.jobSet.jobs.length === 0 &&
      data?.jobSet.id === globalIdInput
    ) {
      !loading &&
        setError('globalId', {
          type: 'jobsetError',
          message: 'This Global ID does not exist',
        });
    }
  });

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
