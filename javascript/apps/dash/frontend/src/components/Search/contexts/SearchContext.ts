import noop from 'lodash/noop';
import {createContext} from 'react';

type SearchContextType = {
  isOpen: boolean;
  setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
  searchValue: string;
  setSearchValue: React.Dispatch<React.SetStateAction<string>>;
  debouncedValue: string;
  reset: () => void;
  history: string[];
  setHistory: React.Dispatch<React.SetStateAction<string[]>>;
};

export default createContext<SearchContextType>({
  isOpen: false,
  setIsOpen: noop,
  searchValue: '',
  setSearchValue: noop,
  debouncedValue: '',
  reset: noop,
  history: [''],
  setHistory: noop,
});
