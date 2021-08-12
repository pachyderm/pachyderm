import classNames from 'classnames';
import React, {useEffect, useRef} from 'react';

import useFadeOut from '../hooks/useFadeOut';
import {CheckmarkSVG} from '../Svg';

import styles from './NotificationBanner.module.css';

export type NotificationBannerProps = {
  duration: number;
  remove: () => void;
  type: 'success' | 'error';
};

const NotificationBanner: React.FC<NotificationBannerProps> = ({
  children,
  remove,
  duration,
  type,
}) => {
  const fadeOut = useFadeOut(duration - 1000);

  const removeRef = useRef(remove);
  removeRef.current = remove;

  useEffect(() => {
    const removeId = setTimeout(() => {
      removeRef.current();
    }, duration);
    return () => {
      clearTimeout(removeId);
    };
  }, [duration]);

  const containerClasses = classNames(styles.base, fadeOut);
  const iconClasses = classNames(styles.icon, {
    [styles.iconError]: type === 'error',
  });

  return (
    <div className={containerClasses}>
      <div className={iconClasses}>
        {type === 'success' ? (
          <CheckmarkSVG
            className={styles.svg}
            data-testid="NotificationBanner__checkmark"
          />
        ) : (
          <span
            className={styles.exclamation}
            data-testid="NotificationBanner__error"
          >
            !
          </span>
        )}
      </div>
      <span role="alert" className={styles.text}>
        {children}
      </span>
    </div>
  );
};

export default NotificationBanner;
