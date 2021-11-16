import React from 'react';

import {DocumentSVG} from 'Svg';

import styles from './Module.module.css';

type ModuleProps = {
  title: string;
};

const Module: React.FC<ModuleProps> = ({children, title}) => {
  return (
    <div className={styles.base}>
      <div className={styles.title}>
        <DocumentSVG className={styles.documentSvg} />
        {title}
      </div>
      {children}
    </div>
  );
};

export default Module;
