import {useContext} from 'react';

import SelectedIdContext from 'Dropdown/contexts/SelectedIdContext';

const useSelectedId = () => {
  return useContext(SelectedIdContext);
};

export default useSelectedId;
