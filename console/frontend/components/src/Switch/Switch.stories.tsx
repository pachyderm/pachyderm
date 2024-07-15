import React from 'react';

import {Switch} from './';

export default {title: 'Switch'};

export const Default = () => {
  return <Switch />;
};

export const Checked = () => {
  return <Switch defaultChecked={true} />;
};
