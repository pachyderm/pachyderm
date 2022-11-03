import React from 'react';

type VariationProps = {
  value: boolean;
};

const Variation: React.FC<VariationProps> = ({children}) => {
  return <>{children}</>;
};

export default Variation;
