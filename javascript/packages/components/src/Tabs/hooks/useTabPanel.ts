import {useContext, useMemo, useCallback} from 'react';

import TabsContext from '../TabsContext';

const useTabPanel = (id: string) => {
  const {activeTabId, setActiveTabId} = useContext(TabsContext);

  const isShown = useMemo(() => id === activeTabId, [activeTabId, id]);
  const selectTab = useCallback(
    (e: React.SyntheticEvent) => {
      e.preventDefault();

      setActiveTabId(id);
    },
    [setActiveTabId, id],
  );

  return {isShown, selectTab, setActiveTabId};
};

export default useTabPanel;
