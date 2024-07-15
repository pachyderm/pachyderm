import React from 'react';

import EllipsisMessage from './EllipsisMessage';

export default {
  title: 'EllipsisMessage',
  parameters: {
    layout: 'centered',
  },
};

export const Default = () => {
  return (
    <div style={{width: '4rem'}}>
      <EllipsisMessage message="This is text that needs a tooltip." />
    </div>
  );
};
