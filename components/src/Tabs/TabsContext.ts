import noop from 'lodash/noop';
import {createContext} from 'react';

export interface TabsContextInterface {
  activeTabId: string;
  setActiveTabId: (nextState: string) => void;
}

const TabsContext = createContext<TabsContextInterface>({
  activeTabId: '',
  setActiveTabId: noop,
});

export default TabsContext;
