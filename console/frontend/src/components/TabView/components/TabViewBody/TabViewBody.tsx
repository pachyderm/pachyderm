import React from 'react';

import {Tabs, SkeletonBodyText} from '@pachyderm/components';

import BodyContent from './components/BodyContent';
import BodyHeader from './components/BodyHeader';
import BodyHeaderDropdown from './components/BodyHeaderDropdown';
import BodyHeaderTabs from './components/BodyHeaderTabs';
import styles from './TabViewBody.module.css';

export type TabViewBodyProps = {
  children?: React.ReactNode;
  initialActiveTabId: string;
  showSkeleton: boolean;
  onSwitch?: (activeTab: string) => void;
};

const TabViewBody: React.FC<TabViewBodyProps> = ({
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
          data-testid="TabViewBody__loadingIndicator"
        />
      ) : (
        <Tabs initialActiveTabId={initialActiveTabId} onSwitch={onSwitch}>
          {children}
        </Tabs>
      )}
    </div>
  );
};

export default Object.assign(TabViewBody, {
  Header: BodyHeader,
  Dropdown: BodyHeaderDropdown,
  Tabs: BodyHeaderTabs,
  Content: BodyContent,
});
