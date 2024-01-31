import classnames from 'classnames';
import React from 'react';

import styles from './Header.module.css';

type HeaderProps = {
  appearance?: 'light' | 'dark';
  hasSubheader?: boolean;
  children?: React.ReactNode;
  className?: string;
};

const Header: React.FC<HeaderProps> = ({
  children,
  appearance = 'dark',
  hasSubheader = false,
  className,
}) => {
  return (
    <header
      className={classnames(styles.base, className, {
        [styles[appearance]]: true,
        [styles['hasSubheader']]: hasSubheader,
      })}
    >
      {children}
    </header>
  );
};

export default Header;
