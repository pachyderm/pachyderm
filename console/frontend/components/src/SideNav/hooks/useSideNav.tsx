import {useContext, useCallback} from 'react';

import sideNavContext from '../SideNavContext';

const useSideNav = () => {
  const {minimized, setMinimized} = useContext(sideNavContext);

  const toggleMinimized = useCallback(() => {
    setMinimized(!minimized);
  }, [setMinimized, minimized]);

  return {minimized, toggleMinimized};
};

export default useSideNav;
