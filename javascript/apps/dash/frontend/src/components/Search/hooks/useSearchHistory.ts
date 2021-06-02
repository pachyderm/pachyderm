import omit from 'lodash/omit';
import {useCallback, useEffect, useState} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
interface SearchHistory {
  [key: string]: string[];
}

export const useSearchHistory = () => {
  const [fullHistory, setFullHistory] = useState<SearchHistory>({});
  const {projectId} = useUrlState();

  useEffect(() => {
    const retrievedData = window.localStorage.getItem('searchHistory');
    if (retrievedData) {
      setFullHistory(JSON.parse(retrievedData));
    }
  }, [projectId]);

  useEffect(() => {
    window.localStorage.setItem('searchHistory', JSON.stringify(fullHistory));
  }, [fullHistory, projectId]);

  const history = fullHistory[projectId] ? fullHistory[projectId] : [];

  const setHistory = useCallback(
    (newHistory: string[]) => {
      if (newHistory.length === 0) {
        const tempHistory = omit(fullHistory, projectId);
        setFullHistory(tempHistory);
      } else {
        setFullHistory({...fullHistory, [projectId]: newHistory});
      }
    },
    [fullHistory, projectId],
  );

  return {
    history,
    setHistory,
  };
};
