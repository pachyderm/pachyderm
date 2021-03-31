import React from 'react';
import {render} from 'react-dom';

import '@pachyderm/polyfills';
import '@pachyderm/components/dist/style.css';

import 'styles/index.css';

import DashUI from './DashUI';

render(<DashUI />, document.getElementById('root'));
