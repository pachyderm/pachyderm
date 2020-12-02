import React from 'react';
import {render} from 'react-dom';

// TODO: move these to base polyfill library
import 'focus-visible';
import 'url-search-params-polyfill';

import 'styles/index.css';

import DashUI from './DashUI';

render(<DashUI />, document.getElementById('root'));
