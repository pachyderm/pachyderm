import {useEffect} from 'react';
import Splitting from 'splitting';

const useSplitting = () => {
  useEffect(() => {
    Splitting();
  }, []);

  const props = {
    'data-splitting': '',
  };

  return props;
};

export default useSplitting;
