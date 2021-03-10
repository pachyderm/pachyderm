import React from 'react';

import {SkeletonBodyText} from 'SkeletonBodyText';
import {Tabs} from 'Tabs';

import BodyContent from './components/BodyContent';
import BodyHeader from './components/BodyHeader';
import BodyHeaderDropdown from './components/BodyHeaderDropdown';
import BodyHeaderTabs from './components/BodyHeaderTabs';
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

export default Object.assign(TableViewBody, {
  Header: BodyHeader,
  Dropdown: BodyHeaderDropdown,
  Tabs: BodyHeaderTabs,
  Content: BodyContent,
});
