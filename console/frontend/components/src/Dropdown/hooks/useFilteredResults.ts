import {useContext} from 'react';

import FilteredResultsContext from '../contexts/FilteredResultsContext';

const useFilteredResults = () => {
  return useContext(FilteredResultsContext);
};

export default useFilteredResults;
