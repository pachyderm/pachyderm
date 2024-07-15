import React from 'react';

import {Tabs} from '@pachyderm/components';

import styles from './TableViewSection.module.css';

type TableViewSectionProps = {
  children?: React.ReactNode;
  id: string;
};

const TableView: React.FC<TableViewSectionProps> = ({id, children}) => {
  return (
    <Tabs.TabPanel id={id} className={styles.base} renderWhenHidden={false}>
      {children}
    </Tabs.TabPanel>
  );
};

export default TableView;
