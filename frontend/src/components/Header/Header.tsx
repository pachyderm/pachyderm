import classnames from 'classnames';
import React from 'react';

import styles from './Header.module.css';

type HeaderProps = {
  appearance?: 'light' | 'dark';
  children?: React.ReactNode;
};

const Header: React.FC<HeaderProps> = ({children, appearance = 'dark'}) => {
  return (
    <header className={classnames(styles.base, {[styles[appearance]]: true})}>
      {children}
    </header>
  );
};

export default Header;
