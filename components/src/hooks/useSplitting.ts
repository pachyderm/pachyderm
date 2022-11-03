import {useEffect} from 'react';
import Splitting from 'splitting';

const useSplitting = () => {
  useEffect(() => {
    Splitting();
  }, []);

  const props = {
    // eslint-disable-next-line @typescript-eslint/naming-convention
    'data-splitting': '',
  };

  return props;
};

export default useSplitting;
