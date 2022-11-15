import noop from 'lodash/noop';
import {createContext} from 'react';

const sideNavContext = createContext({
  minimized: false,
  setMinimized: noop,
});

export default sideNavContext;
