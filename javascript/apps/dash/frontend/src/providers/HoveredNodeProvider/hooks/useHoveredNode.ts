import {useContext} from 'react';

import HoveredNodeContext from '../contexts/HoveredNodeContext';

const useHoveredNode = () => {
  const {hoveredNode, setHoveredNode} = useContext(HoveredNodeContext);

  return {
    hoveredNode,
    setHoveredNode,
  };
};

export default useHoveredNode;
