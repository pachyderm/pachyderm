import {useState, useEffect} from 'react';

import styles from './useFadeOut.module.css';

const useFadeOut = (delay: number) => {
  const [fadeOut, setFadeOut] = useState(false);

  useEffect(() => {
    const id = setTimeout(() => setFadeOut(true), delay);
    return () => clearTimeout(id);
  }, [delay]);

  return fadeOut ? styles.fadeOut : '';
};

export default useFadeOut;
