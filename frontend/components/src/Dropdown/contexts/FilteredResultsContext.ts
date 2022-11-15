import {createContext} from 'react';

import {ItemObject} from '../Dropdown';

export interface IFilteredResultsContext {
  filteredResults: ItemObject[];
}

export default createContext<IFilteredResultsContext>({
  filteredResults: [],
});
