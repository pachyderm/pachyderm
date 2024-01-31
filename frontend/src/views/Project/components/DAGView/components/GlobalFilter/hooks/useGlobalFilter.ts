import {useCallback, useState, useRef, useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import {useJobSet, useJobSetLazy} from '@dash-frontend/hooks/useJobSet';
import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  getStandardDateFromISOString,
  getUnixSecondsFromISOString,
} from '@dash-frontend/lib/dateTime';
import {InternalJobSet, NodeState} from '@dash-frontend/lib/types';
import {useOutsideClick} from '@pachyderm/components';

type GlobalIdFilterFormValues = {
  globalId: string;
  filter?: string;
  startDate: string;
  endDate: string;
  startTime: string;
  endTime: string;
};

export const FILTER = 'filter';
export const FILTER_HOUR = 'lastHour';
export const FILTER_DAY = 'lastDay';
export const FILTER_WEEK = 'lastWeek';

const useGlobalFilter = () => {
  const {searchParams, updateSearchParamsAndGo, getNewSearchParamsAndGo} =
    useUrlQueryState();
  const {projectId} = useUrlState();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [dateTimePickerOpen, setDateTimePickerOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const {
    jobs,
    loading: jobsLoading,
    error: jobsError,
  } = useJobSets(projectId, 20);

  const globalIdFilter = searchParams.globalIdFilter;
  const formCtx = useForm<GlobalIdFilterFormValues>({
    mode: 'onChange',
    defaultValues: {
      globalId: globalIdFilter,
      filter: undefined,
    },
  });

  const {jobSet: jobSets} = useJobSet(globalIdFilter, Boolean(globalIdFilter));
  const selectedJobSet = useMemo(() => {
    if (jobSets?.length && jobSets?.[0]?.job?.id === globalIdFilter) {
      return jobSets[0];
    }
    return undefined;
  }, [globalIdFilter, jobSets]);

  const buttonText = useMemo(
    () =>
      selectedJobSet?.created
        ? `Viewing ${getStandardDateFromISOString(selectedJobSet.created)}`
        : null,
    [selectedJobSet],
  );

  const {watch, setError, clearErrors, reset, formState, setValue, register} =
    formCtx;
  const globalIdInput = watch('globalId');

  const removeFilter = useCallback(() => {
    clearErrors('globalId');
    setValue('globalId', '');
    getNewSearchParamsAndGo({
      globalIdFilter: undefined,
    });
  }, [clearErrors, getNewSearchParamsAndGo, setValue]);

  const {jobSet, loading, ...rest} = useJobSetLazy(globalIdInput);

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
    }
  }, [jobSet, loading, setError, updateSearchParamsAndGo]);

  const handleClick = (jobSetId: string) => {
    updateSearchParamsAndGo({globalIdFilter: jobSetId});
  };

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

  const isLastHour = (started?: string) => {
    const hourInSeconds = 60 * 60;
    const now = Math.floor(Date.now() / 1000);
    const unixStarted = getUnixSecondsFromISOString(started);
    return now - unixStarted <= hourInSeconds;
  };
  const isLastDay = (started?: string) => {
    const dayInSeconds = 60 * 60 * 24;
    const now = Math.floor(Date.now() / 1000);
    const unixStarted = getUnixSecondsFromISOString(started);
    return now - unixStarted <= dayInSeconds;
  };
  const isLastWeek = (started?: string) => {
    const weekInSeconds = 60 * 60 * 24 * 7;
    const now = Math.floor(Date.now() / 1000);
    const unixStarted = getUnixSecondsFromISOString(started);
    return now - unixStarted <= weekInSeconds;
  };

  const chips: {lastHour: number; lastDay: number; lastWeek: number} =
    useMemo(() => {
      const chips = {lastHour: 0, lastDay: 0, lastWeek: 0};
      jobs?.forEach((job) => {
        if (
          isLastHour(job.created || '') &&
          restJobStateToNodeState(job.state) === NodeState.ERROR
        ) {
          chips.lastHour += 1;
        }
        if (
          isLastDay(job.created || '') &&
          restJobStateToNodeState(job.state) === NodeState.ERROR
        ) {
          chips.lastDay += 1;
        }
        if (
          isLastWeek(job.created || '') &&
          restJobStateToNodeState(job.state) === NodeState.ERROR
        ) {
          chips.lastWeek += 1;
        }
      });
      return chips;
    }, [jobs]);

  const clearFilters = () => {
    setValue('filter', undefined);
  };

  const filter = watch(FILTER);

  const toggleDateTimePicker = () => {
    setDateTimePickerOpen((val) => !val);
  };

  const startDate = watch('startDate');
  const endDate = watch('endDate');
  const startTime = watch('startTime');
  const endTime = watch('endTime');

  const dateTimeFilterValue = useMemo(() => {
    if (startDate && endDate) {
      return `${new Date(
        `${startDate}T${startTime ? startTime : '00:00'}:00`,
      ).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        hourCycle: 'h23',
      })} - ${new Date(
        `${endDate}T${endTime ? endTime : '23:59'}:59`,
      ).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric',
        hourCycle: 'h23',
      })}`;
    } else {
      return null;
    }
  }, [startDate, endDate, startTime, endTime]);

  const filteredJobs = useMemo(() => {
    // simple filter
    if (!dateTimePickerOpen) {
      let jobFilter: (job: InternalJobSet) => boolean;
      if (filter === FILTER_HOUR) {
        jobFilter = (job: InternalJobSet) =>
          isLastHour(job.created) &&
          restJobStateToNodeState(job.state) === NodeState.ERROR;
      } else if (filter === FILTER_DAY) {
        jobFilter = (job) =>
          isLastDay(job.created) &&
          restJobStateToNodeState(job.state) === NodeState.ERROR;
      } else {
        jobFilter = (job) =>
          isLastWeek(job.created) &&
          restJobStateToNodeState(job.state) === NodeState.ERROR;
      }

      return filter ? jobs?.filter((job) => jobFilter(job)) : jobs;
    } else {
      // use advanced datetime filter
      const isWithinDateTime = (started: string) => {
        const start = new Date(
          `${startDate}T${startTime ? startTime : '00:00'}:00`,
        );
        const unixStartDate = start.getTime() / 1000;

        const end = new Date(`${endDate}T${endTime ? endTime : '23:59'}:59`);
        const unixEndDate = end.getTime() / 1000;

        const unixStarted = getUnixSecondsFromISOString(started);

        return unixStarted >= unixStartDate && unixStarted <= unixEndDate;
      };

      return startDate && endDate
        ? jobs?.filter((job) => isWithinDateTime(job.created || ''))
        : jobs;
    }
  }, [
    dateTimePickerOpen,
    endDate,
    endTime,
    filter,
    jobs,
    startDate,
    startTime,
  ]);

  return {
    containerRef,
    formCtx,
    dropdownOpen,
    setDropdownOpen,
    loading,
    globalIdFilter,
    globalIdInput,
    handleClick,
    jobsCount: jobs?.length,
    filteredJobs,
    jobsLoading,
    jobsError,
    chips,
    clearFilters,
    showClearButton: typeof filter !== 'undefined',
    dateTimePickerOpen,
    toggleDateTimePicker,
    register,
    dateTimeFilterValue,
    buttonText,
    removeFilter,
  };
};

export default useGlobalFilter;
