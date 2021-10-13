import React from 'react';

import {Switch} from './';

/* eslint-disable-next-line import/no-anonymous-default-export */
export default {title: 'Switch'};

export const Default = () => {
  return <Switch />;
};

export const Checked = () => {
  return <Switch defaultChecked={true} />;
};
