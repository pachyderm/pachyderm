import noop from 'lodash/noop';
import {createContext} from 'react';

export default createContext({
  isOpen: false,
  setIsOpen: noop,
  selectedId: '',
  setSelectedId: noop,
  onSelect: noop,
});
