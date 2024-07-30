import noop from 'lodash/noop';
import {useEffect, useState} from 'react';

import styles from './usePopUp.module.css';

const usePopUp = (show = true) => {
  const [showing, setShowing] = useState(false);
  const [fading, setFading] = useState(false);

  useEffect(() => {
    if (show) {
      setFading(false);
      setShowing(true);
      return () => noop;
    } else {
      setFading(true);
      const id = setTimeout(() => setShowing(false), 500);
      return () => clearTimeout(id);
    }
  }, [show]);

  return {
    animation: fading ? styles.hide : styles.show,
    showing,
  };
};

export default usePopUp;
