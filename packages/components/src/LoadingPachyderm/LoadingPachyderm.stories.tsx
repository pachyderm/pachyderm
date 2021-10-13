import React from 'react';

import {LoadingPachyderm} from '../LoadingPachyderm';

/* eslint-disable-next-line import/no-anonymous-default-export */
export default {title: 'Loading Workspace Animation'};

export const Default = () => {
  return <LoadingPachyderm />;
};

export const FreeTrial = () => {
  return <LoadingPachyderm freeTrial={true} />;
};
