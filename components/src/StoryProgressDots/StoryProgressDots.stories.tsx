import React from 'react';

import StoryProgressDots from './StoryProgressDots';

export default {title: 'StoryProgressDots'};

export const Dark = () => {
  return (
    <div style={{backgroundColor: 'black', padding: '1rem'}}>
      <StoryProgressDots progress={2} stories={6} dotStyle="dark" />
    </div>
  );
};

export const Light = () => {
  return <StoryProgressDots progress={2} stories={6} dotStyle="light" />;
};
