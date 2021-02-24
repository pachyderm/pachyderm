import noop from 'lodash/noop';
import React, {useEffect, useMemo, useState} from 'react';

import Tab from './components/Tab';
import TabPanel from './components/TabPanel';
import TabsHeader from './components/TabsHeader';
import TabsContext from './TabsContext';

export interface TabsProps {
  initialActiveTabId: string;
  onSwitch?: (activeTab: string) => void;
}

const Tabs: React.FC<TabsProps> = ({
  children,
  initialActiveTabId,
  onSwitch = noop,
}) => {
  const [activeTabId, setActiveTabId] = useState<string>(initialActiveTabId);

  useEffect(() => {
    if (activeTabId) onSwitch(activeTabId);
  }, [activeTabId, onSwitch]);

  const contextValue = useMemo(
    () => ({
      activeTabId,
      setActiveTabId,
    }),
    [activeTabId, setActiveTabId],
  );
  return (
    <TabsContext.Provider value={contextValue}>{children}</TabsContext.Provider>
  );
};

export default Object.assign(Tabs, {Tab, TabsHeader, TabPanel});
