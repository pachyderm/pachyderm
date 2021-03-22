import React from 'react';

import styles from './Sidebar.module.css';

const Sidebar: React.FC = ({children}) => (
  <div className={styles.base}>{children}</div>
);

export default Sidebar;
