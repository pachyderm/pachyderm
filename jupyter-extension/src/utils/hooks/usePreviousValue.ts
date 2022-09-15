import {useEffect, useRef} from 'react';

const usePreviousValue = <T>(value: T): T | undefined => {
  const ref = useRef<T>();
  useEffect(() => {
    ref.current = value;
  });
  return ref.current;
};

export default usePreviousValue;
