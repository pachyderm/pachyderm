import React from 'react';

import {DocumentSVG} from '@pachyderm/components';

import styles from './Module.module.css';

type ModuleProps = {
  children?: React.ReactNode;
  title: string;
};

const Module: React.FC<ModuleProps> = ({children, title}) => {
  return (
    <div className={styles.base}>
      <div className={styles.title}>
        <DocumentSVG className={styles.documentSvg} />
        <h6>{title}</h6>
      </div>
      {children}
    </div>
  );
};

export default Module;
