import React from 'react';

import styles from './Header.module.css';

const Header: React.FC = ({children}) => {
  return <header className={styles.base}>{children}</header>;
};

export default Header;
