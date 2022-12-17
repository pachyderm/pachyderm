import React, {useCallback, useEffect, useState} from 'react';

import useDatums from '@dash-frontend/hooks/useDatums';
import useDatumSearch from '@dash-frontend/hooks/useDatumSearch';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DATUM_ID_LENGTH} from '@dash-frontend/views/DatumViewer/constants/DatumViewer';
import useDatumPath from '@dash-frontend/views/DatumViewer/hooks/useDatumPath';

const useDatumList = (
  setIsExpanded: React.Dispatch<React.SetStateAction<boolean>>,
) => {
  const [searchValue, setSearchValue] = useState('');
  const [searchedDatumId, setSearchedDatumId] = useState('');

  const {projectId, pipelineId} = useUrlState();
  const {currentJobId, currentDatumId, updateSelectedDatum} = useDatumPath();

  const {viewState} = useUrlQueryState();

  const isSearchValid = searchValue.length === DATUM_ID_LENGTH;

  useEffect(() => {
    clearSearch();
  }, [currentJobId]);

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

  const {datums, loading} = useDatums({
    projectId,
    pipelineId,
    jobId: currentJobId,
    filter: viewState.datumFilters,
  });

  const displayDatums = searchedDatum ? [searchedDatum] : datums;

  const showNoSearchResults = !searchLoading && isSearchValid && !searchedDatum;
  return {
    datums: displayDatums,
    currentDatumId,
    onDatumClick,
    loading: loading,
    searchValue,
    setSearchValue,
    clearSearch,
    showNoSearchResults,
    searchedDatum,
  };
};

export default useDatumList;
