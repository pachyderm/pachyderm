import {useCallback, useState, useRef, useEffect, useMemo} from 'react';
import {useForm} from 'react-hook-form';

import {jqAnd, jqSelect} from '@dash-frontend/api/utils/jqHelpers';
import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import {useJobSet, useJobSetLazy} from '@dash-frontend/hooks/useJobSet';
import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  getStandardDateFromISOString,
  getUnixSecondsFromISOString,
} from '@dash-frontend/lib/dateTime';
import {NodeState} from '@dash-frontend/lib/types';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  useDebounce,
  useOutsideClick,
  usePreviousValue,
} from '@pachyderm/components';

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
  const {
    searchParams: {globalIdFilter},
    updateSearchParamsAndGo,
    getNewSearchParamsAndGo,
  } = useUrlQueryState();
  const {projectId} = useUrlState();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [dateTimePickerOpen, setDateTimePickerOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Form context
  const formCtx = useForm<GlobalIdFilterFormValues>({
    mode: 'onChange',
    defaultValues: {
      globalId: globalIdFilter,
      filter: undefined,
    },
  });
  const {watch, setError, clearErrors, reset, formState, setValue, register} =
    formCtx;
  const globalIdInput = watch('globalId');
  const startDate = watch('startDate');
  const endDate = watch('endDate');
  const startTime = watch('startTime');
  const endTime = watch('endTime');
  const filter = watch(FILTER);

  // advanced date
  const dateTimeStart = useMemo(
    () => new Date(`${startDate}T${startTime ? startTime : '00:00'}:00`),
    [startDate, startTime],
  );
  const dateTimeEnd = useMemo(
    () => new Date(`${endDate}T${endTime ? endTime : '23:59'}:59`),
    [endDate, endTime],
  );
  let jqFilter = '';

  const jobsListArgs: string[] = [];
  if (dateTimePickerOpen) {
    if (startDate) {
      const unixTime = Math.floor(dateTimeStart.getTime() / 1000);
      const jq = `(.created | (split(".")[0] + "Z") | fromdateiso8601 >= ${unixTime})`;
      jobsListArgs.push(jq);
    }
    if (endDate) {
      const unixTime = Math.floor(dateTimeEnd.getTime() / 1000);
      const jq = `(.created | (split(".")[0] + "Z") | fromdateiso8601 <= ${unixTime})`;
      jobsListArgs.push(jq);
    }

    // Displays the list of job sets
    jqFilter = (startDate || endDate) && jqSelect(jqAnd(...jobsListArgs));
  }

  const {
    jobs,
    loading: jobsLoading,
    error: jobsError,
  } = useJobSets({projectName: projectId, jqFilter}, 100, dropdownOpen);

  const failedChipsArgs: string[] = [];

  const lastSevenDays = new Date(Date.now());
  lastSevenDays.setSeconds(59);
  lastSevenDays.setDate(lastSevenDays.getDate() - 7);
  const time = Math.floor(lastSevenDays.getTime() / 1000);

  const jq = `(.created | (split(".")[0] + "Z") | fromdateiso8601 >= ${time})`;
  failedChipsArgs.push(jq);

  const jqFilter2 = jqSelect(
    jqAnd(
      ...failedChipsArgs,
      '((.state == "JOB_FAILURE") or (.state == "JOB_KILLED") or (.state == "JOB_UNRUNNABLE"))',
    ),
  );

  const {
    jobs: failedJobs,
    loading: _failedJobsLoading,
    error: _failedJobsError,
  } = useJobSets(
    {
      projectName: projectId,
      jqFilter: jqFilter2,
    },
    100,
    dropdownOpen,
  );

  const loadingDebounce = useDebounce(jobsLoading, 1000);

  // This is used grab the date from the selected job
  const {jobSet: jobSets} = useJobSet(globalIdFilter, Boolean(globalIdFilter));
  const headerButtonText = useMemo(
    () =>
      jobSets?.[0]?.created
        ? getStandardDateFromISOString(jobSets[0].created)
        : null,
    [jobSets],
  );

  const clearGlobalIdFilterAndInput = useCallback(() => {
    clearErrors('globalId');
    setValue('globalId', '');
    getNewSearchParamsAndGo({
      globalIdFilter: undefined,
    });
  }, [clearErrors, getNewSearchParamsAndGo, setValue]);

  // If the URL search param updates, set the value to the text box
  // Page load AND when you click a search result
  useEffect(() => {
    if (globalIdFilter) {
      setValue('globalId', globalIdFilter);
    }
  }, [globalIdFilter, setValue]);

  const {loading, refetch} = useJobSetLazy(globalIdInput);

  const updateRoute = (id: string) => {
    const route = lineageRoute({projectId}, false);
    updateSearchParamsAndGo({globalIdFilter: id}, route);
  };

  const handleApplyFilter = async () => {
    const {data: jobSet} = await refetch();
    if (jobSet?.length === 0) {
      setError('globalId', {
        type: 'jobsetError',
        message: 'This Global ID does not exist',
      });
    } else if (jobSet?.length !== 0 && jobSet?.[0]?.job?.id) {
      if (jobSet?.[0]?.job.pipeline?.project?.name === projectId) {
        updateRoute(jobSet[0].job.id);
      } else {
        setError('globalId', {
          type: 'jobsetError',
          message: 'This Global ID from a different project',
        });
      }
    }
  };

  const handleClickSearchResult = (jobSetId: string) => updateRoute(jobSetId);
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
      failedJobs?.forEach((job) => {
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
    }, [failedJobs]);

  const clearFilters = () => {
    setValue('filter', undefined);
  };

  const toggleDateTimePicker = () => {
    setDateTimePickerOpen((val) => !val);
  };

  const dateTimeFilterValue = useMemo(() => {
    if (startDate || endDate) {
      const words: string[] = [];

      if (!endDate) {
        words.push('From:');
      }

      if (startDate) {
        words.push(
          dateTimeStart.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            hourCycle: 'h23',
          }),
        );
      }

      if (startDate && endDate) {
        words.push('-');
      }

      if (!startDate) {
        words.push('To:');
      }

      if (endDate) {
        words.push(
          dateTimeEnd.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: 'numeric',
            minute: 'numeric',
            hourCycle: 'h23',
          }),
        );
      }

      return words.join(' ');
    } else {
      return null;
    }
  }, [startDate, endDate, dateTimeStart, dateTimeEnd]);

  const showApplyFilterButton =
    !formState.errors.globalId &&
    Boolean(globalIdInput) &&
    globalIdInput !== globalIdFilter;

  const prevJobs = usePreviousValue(jobs);
  const memoizedJobs = useMemo(() => {
    return jobsLoading ? prevJobs : jobs;
  }, [jobs, jobsLoading, prevJobs]);
  const showClearButton = typeof filter !== 'undefined';

  const filteredJobs = useMemo(() => {
    // simple filter
    if (showClearButton) {
      return failedJobs?.filter((job) => {
        if (filter === FILTER_HOUR) {
          return isLastHour(job.created);
        } else if (filter === FILTER_DAY) {
          return isLastDay(job.created);
        } else {
          return isLastWeek(job.created);
        }
      });
    } else {
      // advanced filter
      return memoizedJobs;
    }
  }, [failedJobs, filter, memoizedJobs, showClearButton]);

  return {
    containerRef,
    formCtx,
    dropdownOpen,
    setDropdownOpen,
    loading,
    globalIdFilter,
    globalIdInput,
    handleClickSearchResult,
    jobsCount: jobs?.length,
    filteredJobs,
    jobsLoading: loadingDebounce,
    jobsError,
    chips,
    clearFilters,
    showClearButton,
    dateTimePickerOpen,
    toggleDateTimePicker,
    register,
    dateTimeFilterValue,
    headerButtonText,
    clearGlobalIdFilterAndInput,
    showApplyFilterButton,
    handleApplyFilter,
  };
};

export default useGlobalFilter;
