import noop from 'lodash/noop';
import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {useHistory, useRouteMatch} from 'react-router-dom';

import generatePathWithSearch from 'lib/generatePathWithSearch';

import Tab from './components/Tab';
import TabPanel from './components/TabPanel';
import TabsHeader from './components/TabsHeader';
import TabsContext from './TabsContext';

export interface StatefulTabsProps {
  initialActiveTabId: string;
  onSwitch?: (activeTab: string) => void;
}

const StatefulTabs: React.FC<StatefulTabsProps> = ({
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

interface RouterTabsProps
  extends Omit<StatefulTabsProps, 'initialActiveTabId'> {
  basePath: string;
  basePathTabId: string;
  tabIdParam?: string;
}

const RouterTabs: React.FC<RouterTabsProps> = ({
  children,
  onSwitch = noop,
  basePath,
  basePathTabId,
  tabIdParam = 'tabId',
}) => {
  const match = useRouteMatch<{[key: string]: string}>(basePath);
  const browserHistory = useHistory();

  const activeTabId = useMemo(() => {
    if (!match) {
      return '';
    }

    return match.params[tabIdParam] || basePathTabId;
  }, [match, tabIdParam, basePathTabId]);

  const setActiveTabId = useCallback(
    (tabId: string) => {
      browserHistory.push(
        generatePathWithSearch(basePath, {
          ...(match?.params || {}),
          [tabIdParam]: tabId,
        }),
      );
    },
    [browserHistory, basePath, tabIdParam, match],
  );

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

export default Object.assign(StatefulTabs, {
  Tab,
  TabsHeader,
  TabPanel,
  RouterTabs,
});
