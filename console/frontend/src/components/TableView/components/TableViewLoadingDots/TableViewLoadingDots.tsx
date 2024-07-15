import React from 'react';

import {LoadingDots} from '@pachyderm/components';

import styles from './TableViewLoadingDots.module.css';

type TableViewSectionProps = {
  ['data-testid']: string;
};

const TableView: React.FC<TableViewSectionProps> = ({
  'data-testid': dataTestId,
}) => {
  return (
    <div className={styles.base} data-testid={dataTestId}>
      <LoadingDots />
    </div>
  );
};

export default TableView;
