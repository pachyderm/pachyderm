import noop from 'lodash/noop';
import {createContext} from 'react';

export default createContext({
  hoveredNode: '',
  setHoveredNode: noop,
});
