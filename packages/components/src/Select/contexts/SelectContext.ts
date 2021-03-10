import noop from 'lodash/noop';
import {createContext} from 'react';

import {OptionRef} from '../lib/types';

interface SelectContext {
  id: string;
  value: string;
  activeValue: string;
  activeOptionRef: React.MutableRefObject<HTMLElement | null> | null;
  displayValue: React.ReactNode;
  activeIndex?: number;
  options: OptionRef[];
  isOpen: boolean;
  ignoreBlur: boolean;
  selectOptionAtIndex: (index: number) => void;
  setActiveValue: React.Dispatch<React.SetStateAction<string>>;
  setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
  setOptions: React.Dispatch<React.SetStateAction<OptionRef[]>>;
  setActiveIndex: React.Dispatch<React.SetStateAction<number | undefined>>;
  setIgnoreBlur: React.Dispatch<React.SetStateAction<boolean>>;
}

export default createContext<SelectContext>({
  id: '',
  value: '',
  activeValue: '',
  activeOptionRef: null,
  displayValue: '',
  options: [],
  isOpen: false,
  ignoreBlur: false,
  selectOptionAtIndex: noop,
  setActiveValue: noop,
  setIsOpen: noop,
  setOptions: noop,
  setActiveIndex: noop,
  setIgnoreBlur: noop,
});
