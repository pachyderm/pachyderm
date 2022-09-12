import React from 'react';

const LoadingDots: React.FC = () => {
  return (
    <div className={'jp-dots-container'}>
      <div className={'jp-dots-base'} role="status" aria-label="loading" />
    </div>
  );
};

export default LoadingDots;
