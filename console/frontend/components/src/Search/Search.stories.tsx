import noop from 'lodash/noop';
import React from 'react';

import {Search} from './';

export default {title: 'Search'};

export const Default = () => {
  return <Search onSearch={noop} />;
};
