import noop from 'lodash/noop';
import React, {createContext} from 'react';

import {DropdownProps, ItemObject} from 'Dropdown/Dropdown';

export interface IDropdownContext {
  isOpen: boolean;
  setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
  setFilteredResults: React.Dispatch<React.SetStateAction<ItemObject[]>>;
  filter: Required<DropdownProps>['filter'];
  sideOpen: boolean;
}

export default createContext<IDropdownContext>({
  isOpen: false,
  setIsOpen: noop,
  setFilteredResults: noop,
  filter: () => true,
  sideOpen: false,
});
