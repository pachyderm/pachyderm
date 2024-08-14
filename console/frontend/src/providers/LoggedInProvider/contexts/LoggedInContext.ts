import noop from 'lodash/noop';
import {Dispatch, SetStateAction, createContext} from 'react';

interface LoggedInContext {
  loggedIn: boolean;
  setLoggedIn: Dispatch<SetStateAction<boolean>>;
}

export default createContext<LoggedInContext>({
  loggedIn: false,
  setLoggedIn: noop,
});
