import noop from 'lodash/noop';
import React, {createContext} from 'react';

export interface TabsContextInterface {
  activeTabId: string;
  setActiveTabId: React.Dispatch<React.SetStateAction<string>>;
}

const TabsContext = createContext<TabsContextInterface>({
  activeTabId: '',
  setActiveTabId: noop,
});

export default TabsContext;
