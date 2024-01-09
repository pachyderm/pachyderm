import React, {useCallback, useEffect, useState, useMemo} from 'react';

import {usePreviousValue} from '@dash-frontend/../components/src';
import {DatumState} from '@dash-frontend/api/pps';
import {useDatum} from '@dash-frontend/hooks/useDatum';
import {useDatumsPaged} from '@dash-frontend/hooks/useDatums';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DatumFilter} from '@dash-frontend/lib/types';
import {
  DATUM_ID_LENGTH,
  DATUM_LIST_PAGE_SIZE,
} from '@dash-frontend/views/DatumViewer/constants/DatumViewer';
import useDatumPath from '@dash-frontend/views/DatumViewer/hooks/useDatumPath';

type DatumFilterObject = {
  [key in DatumFilter]: number;
};

const useDatumList = (
  setIsExpanded: React.Dispatch<React.SetStateAction<boolean>>,
) => {
  const {projectId, pipelineId} = useUrlState();
  const {searchParams} = useUrlQueryState();
  const {
    urlJobId: currentJobId,
    currentDatumId,
    updateSelectedDatum,
  } = useDatumPath();

  const [searchValue, setSearchValue] = useState('');
  const [searchedDatumId, setSearchedDatumId] = useState('');

  const [page, setPage] = useState(1);
  const [cursors, setCursors] = useState<string[]>(['']);
  const [currentCursor, setCurrentCursor] = useState('');

  const isSearchValid = searchValue.length === DATUM_ID_LENGTH;

  const clearSearch = () => {
    setSearchValue('');
    setSearchedDatumId('');
  };

  const onDatumClick = useCallback(
    (datumId: string) => {
      updateSelectedDatum(currentJobId, datumId);
      setIsExpanded(false);
    },
    [currentJobId, setIsExpanded, updateSelectedDatum],
  );

  useEffect(() => {
    if (isSearchValid) {
      setSearchedDatumId(searchValue);
    }
  }, [isSearchValid, searchValue]);

  const {datum: searchedDatum, loading: searchLoading} = useDatum(
    {
      datum: {
        id: searchedDatumId,
        job: {
          id: currentJobId,
          pipeline: {
            name: pipelineId,
            project: {
              name: projectId,
            },
          },
        },
      },
    },
    isSearchValid,
  );

  const {job} = useJob({
    id: currentJobId,
    pipelineName: pipelineId,
    projectId,
  });

  const isProcessing = !job?.finished;
  const wasProcessing = usePreviousValue(isProcessing);
  const datumMetrics = useMemo<DatumFilterObject>(() => {
    return {
      [DatumState.SUCCESS]: Number(job?.dataProcessed) || 0,
      [DatumState.SKIPPED]: Number(job?.dataSkipped) || 0,
      [DatumState.FAILED]: Number(job?.dataFailed) || 0,
      [DatumState.RECOVERED]: Number(job?.dataRecovered) || 0,
    };
  }, [job]);

  const totalDatums = useMemo(() => {
    const filters = searchParams.datumFilters;
    if (filters && filters?.length !== 0) {
      return (
        filters?.reduce((acc, value) => {
          return acc + datumMetrics[value as DatumFilter];
        }, 0) || 0
      );
    }
    return Object.values(datumMetrics).reduce((a, b) => a + b, 0);
  }, [datumMetrics, searchParams.datumFilters]);

  const {datums, cursor, hasNextPage, loading, refetch} = useDatumsPaged({
    job: {
      id: currentJobId,
      pipeline: {
        name: pipelineId,
        project: {
          name: projectId,
        },
      },
    },
    filter: {
      state: searchParams.datumFilters as DatumState[],
    },
    number: String(DATUM_LIST_PAGE_SIZE),
    paginationMarker: currentCursor,
  });

  const refresh = useCallback(() => {
    clearSearch();
    setPage(1);
    setCursors(['']);
    setCurrentCursor('');

    refetch();
  }, [refetch]);

  useEffect(() => {
    refresh();
  }, [refresh, searchParams.datumFilters]);

  useEffect(() => {
    if (wasProcessing && !isProcessing) {
      refresh();
    }
  }, [isProcessing, refresh, wasProcessing]);

  useEffect(() => {
    if (cursor && cursors.length < page) {
      setCurrentCursor(cursor);
      setCursors((arr) => [...arr, cursor]);
    } else {
      setCurrentCursor(cursors[page - 1]);
    }
  }, [cursor, cursors, page]);

  const displayDatums = searchedDatum ? [searchedDatum] : datums;
  const showNoSearchResults = !searchLoading && isSearchValid && !searchedDatum;

  const pageCount =
    job?.finished && totalDatums
      ? Math.ceil(totalDatums / DATUM_LIST_PAGE_SIZE)
      : undefined;

  const contentLength =
    job?.finished && totalDatums
      ? totalDatums
      : (cursors.length - 1) * DATUM_LIST_PAGE_SIZE + (datums?.length ?? 0);

  return {
    datums: displayDatums,
    currentDatumId,
    onDatumClick,
    loading,
    searchValue,
    setSearchValue,
    clearSearch,
    showNoSearchResults,
    searchedDatum,
    page,
    setPage,
    hasNextPage,
    isProcessing,
    pageCount: pageCount,
    contentLength: contentLength,
    refresh,
    isSearchValid,
  };
};

export default useDatumList;
