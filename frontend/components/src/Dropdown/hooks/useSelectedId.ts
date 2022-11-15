import {useContext} from 'react';

import SelectedIdContext from '../contexts/SelectedIdContext';

const useSelectedId = () => {
  return useContext(SelectedIdContext);
};

export default useSelectedId;
