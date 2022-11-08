import noop from 'lodash/noop';
import {createContext} from 'react';

const TutorialModalContext = createContext({
  minimized: false,
  setMinimized: noop,
});

export default TutorialModalContext;
