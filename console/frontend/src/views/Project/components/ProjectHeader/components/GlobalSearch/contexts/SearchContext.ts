import noop from 'lodash/noop';
import {createContext} from 'react';

type SearchContextType = {
  isOpen: boolean;
  setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
  searchValue: string;
  setSearchValue: React.Dispatch<React.SetStateAction<string>>;
  debouncedValue: string;
  history: string[];
  setHistory: (history: string[]) => void;
};

export default createContext<SearchContextType>({
  isOpen: false,
  setIsOpen: noop,
  searchValue: '',
  setSearchValue: noop,
  debouncedValue: '',
  history: [],
  setHistory: noop,
});
