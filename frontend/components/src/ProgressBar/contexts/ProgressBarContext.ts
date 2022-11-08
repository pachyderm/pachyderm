import noop from 'lodash/noop';
import {createContext} from 'react';

type Context = {
  completed: string[];
  visited: string[];
  complete: (id: string) => void;
  visit: (id: string) => void;
  isVertical: boolean;
  clear: () => void;
};

export default createContext<Context>({
  completed: [],
  visited: [],
  complete: noop,
  visit: noop,
  isVertical: false,
  clear: noop,
});
