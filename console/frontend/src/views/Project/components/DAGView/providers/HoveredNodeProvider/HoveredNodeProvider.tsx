import React, {useMemo, useState} from 'react';

import HoveredNodeContext from './HoveredNodeContext';

type HoveredNodeProviderProps = {
  children?: React.ReactNode;
};

const HoveredNodeProvider: React.FC<HoveredNodeProviderProps> = ({
  children,
}) => {
  const [hoveredNode, setHoveredNode] = useState('');

  const value = useMemo(() => ({hoveredNode, setHoveredNode}), [hoveredNode]);

  return (
    <HoveredNodeContext.Provider value={value}>
      {children}
    </HoveredNodeContext.Provider>
  );
};

export default HoveredNodeProvider;
