import {getApolloContext} from '@apollo/client';
import {DatumFilter} from '@graphqlTypes';
import React, {
  useCallback,
  useEffect,
  useState,
  useMemo,
  useContext,
} from 'react';

import {usePreviousValue} from '@dash-frontend/../components/src';
import useDatums from '@dash-frontend/hooks/useDatums';
import useDatumSearch from '@dash-frontend/hooks/useDatumSearch';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
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
  const {viewState} = useUrlQueryState();
  const {currentJobId, currentDatumId, updateSelectedDatum} = useDatumPath();

  const [searchValue, setSearchValue] = useState('');
  const [searchedDatumId, setSearchedDatumId] = useState('');

  const [page, setPage] = useState(1);
  const [cursors, setCursors] = useState<string[]>(['']);
  const [currentCursor, setCurrentCursor] = useState('');

  const isSearchValid = searchValue.length === DATUM_ID_LENGTH;

  const {client} = useContext(getApolloContext());

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

  const {datum: searchedDatum, loading: searchLoading} = useDatumSearch(
    {
      projectId,
      pipelineId,
      jobId: currentJobId,
      id: searchedDatumId,
    },
    {skip: !isSearchValid},
  );

  const {job} = useJob({
    id: currentJobId,
    pipelineName: pipelineId,
    projectId,
  });

  const isProcessing = job && !job.finishedAt;
  const wasProcessing = usePreviousValue(isProcessing);
  const datumMetrics = useMemo<DatumFilterObject>(() => {
    return {
      [DatumFilter.SUCCESS]: job?.dataProcessed ?? 0,
      [DatumFilter.SKIPPED]: job?.dataSkipped ?? 0,
      [DatumFilter.FAILED]: job?.dataFailed ?? 0,
      [DatumFilter.RECOVERED]: job?.dataRecovered ?? 0,
    };
  }, [job]);

  const totalDatums = useMemo(() => {
    const filters = viewState.datumFilters;
    if (filters && filters?.length !== 0) {
      return (
        filters?.reduce((acc, value) => {
          return acc + datumMetrics[value];
        }, 0) || 0
      );
    }
    return Object.values(datumMetrics).reduce((a, b) => a + b, 0);
  }, [datumMetrics, viewState.datumFilters]);

  const {datums, cursor, hasNextPage, loading, refetch} = useDatums({
    projectId,
    pipelineId,
    jobId: currentJobId,
    filter: viewState.datumFilters,
    limit: DATUM_LIST_PAGE_SIZE,
    cursor: currentCursor,
  });

  const refresh = useCallback(() => {
    clearSearch();
    setPage(1);
    setCursors(['']);
    setCurrentCursor('');

    const cleared = client?.cache.evict({
      id: 'ROOT_QUERY',
      fieldName: 'datums',
    });
    if (cleared) {
      client?.cache.gc();
      refetch({
        args: {
          projectId,
          pipelineId,
          jobId: currentJobId,
          filter: viewState.datumFilters,
          limit: DATUM_LIST_PAGE_SIZE,
          cursor: '',
        },
      });
    }
  }, [
    client?.cache,
    currentJobId,
    pipelineId,
    projectId,
    refetch,
    viewState.datumFilters,
  ]);

  useEffect(() => {
    refresh();
  }, [refresh, viewState.datumFilters]);

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
    job?.finishedAt && totalDatums
      ? Math.ceil(totalDatums / DATUM_LIST_PAGE_SIZE)
      : undefined;

  const contentLength =
    job?.finishedAt && totalDatums
      ? totalDatums
      : (cursors.length - 1) * DATUM_LIST_PAGE_SIZE + datums.length;

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
  };
};

export default useDatumList;
