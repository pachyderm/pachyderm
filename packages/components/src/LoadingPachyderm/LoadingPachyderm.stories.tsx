import React from 'react';

import {LoadingPachyderm} from '../LoadingPachyderm';

export default {title: 'Loading Workspace Animation'};

export const Default = () => {
  return <LoadingPachyderm />;
};

export const FreeTrial = () => {
  return <LoadingPachyderm freeTrial={true} />;
};
