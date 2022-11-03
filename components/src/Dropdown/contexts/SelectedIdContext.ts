import noop from 'lodash/noop';
import {createContext} from 'react';

export interface ISelectedIdContext {
  selectedId: string;
  setSelectedId: (id: string) => void;
}

export default createContext<ISelectedIdContext>({
  selectedId: '',
  setSelectedId: noop,
});
