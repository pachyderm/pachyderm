import {useEffect, useState} from 'react';

export const useSearchHistory = () => {
  const [history, setHistory] = useState<string[]>([]);

  useEffect(() => {
    const retrievedData = window.localStorage.getItem('searchHistory');
    if (retrievedData) {
      setHistory(JSON.parse(retrievedData));
    }
  }, []);

  useEffect(() => {
    window.localStorage.setItem('searchHistory', JSON.stringify(history));
  }, [history]);

  return {
    history,
    setHistory,
  };
};
