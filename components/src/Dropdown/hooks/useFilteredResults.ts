import {useContext} from 'react';

import FilteredResultsContext from 'Dropdown/contexts/FilteredResultsContext';

const useFilteredResults = () => {
  return useContext(FilteredResultsContext);
};

export default useFilteredResults;
