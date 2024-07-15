import React from 'react';

import {Tabs} from '@pachyderm/components';

import styles from './BodyHeader.module.css';

const BodyHeader = ({children}: {children?: React.ReactNode}) => {
  return <Tabs.TabsHeader className={styles.base}>{children}</Tabs.TabsHeader>;
};

export default BodyHeader;
