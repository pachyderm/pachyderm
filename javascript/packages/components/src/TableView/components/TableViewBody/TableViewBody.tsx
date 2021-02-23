import React from 'react';

import SkeletonBodyText from 'SkeletonBodyText';
import Tabs from 'Tabs';

import styles from './TableViewBody.module.css';

export type TableViewBodyProps = {
  initialActiveTabId: string;
  showSkeleton: boolean;
  onSwitch?: (activeTab: string) => void;
};

const TableViewBody: React.FC<TableViewBodyProps> = ({
  initialActiveTabId,
  showSkeleton,
  children,
  onSwitch,
}) => {
  return (
    <div className={styles.base}>
      {showSkeleton ? (
        <SkeletonBodyText
          lines={5}
          data-testid="TableViewBody__loadingIndicator"
        />
      ) : (
        <Tabs initialActiveTabId={initialActiveTabId} onSwitch={onSwitch}>
          {children}
        </Tabs>
      )}
    </div>
  );
};

export default TableViewBody;
