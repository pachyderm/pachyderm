import React, {useMemo, useState} from 'react';

import HoveredNodeContext from './contexts/HoveredNodeContext';

const HoveredNodeProvider: React.FC = ({children}) => {
  const [hoveredNode, setHoveredNode] = useState('');

  const value = useMemo(() => ({hoveredNode, setHoveredNode}), [hoveredNode]);

  return (
    <HoveredNodeContext.Provider value={value}>
      {children}
    </HoveredNodeContext.Provider>
  );
};

export default HoveredNodeProvider;
