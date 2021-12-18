/* eslint-disable @typescript-eslint/no-explicit-any */
import {generateIdTokenForAccount} from '@dash-backend/testHelpers';
import {Account} from '@graphqlTypes';
import {act} from '@testing-library/react';
import userEvent, {ITypeOpts, TargetElement} from '@testing-library/user-event';
import React, {ReactElement} from 'react';
import {BrowserRouter, Route} from 'react-router-dom';

import ApolloProvider from '@dash-frontend/providers/ApolloProvider';

import useRouteController from './hooks/useRouteController';
import {UrlState} from './hooks/useUrlQueryState';
import {Dag} from './lib/types';
import LoggedInProvider from './providers/LoggedInProvider';
import {PROJECT_PATHS} from './views/Project/constants/projectPaths';

export {default as server} from '@dash-backend/index';
export {default as mockServer} from '@dash-backend/mock';

export const withContextProviders = (
  Component: React.ElementType,
): ((props: any) => ReactElement) => {
  // eslint-disable-next-line react/display-name
  return (props: any): any => {
    return (
      <BrowserRouter>
        <LoggedInProvider>
          <ApolloProvider>
            <Component {...props} />
          </ApolloProvider>
        </LoggedInProvider>
      </BrowserRouter>
    );
  };
};

export const MockDAG: React.FC<{dag: Dag}> = ({dag}) => {
  const {selectedNode, navigateToNode} = useRouteController();

  return (
    <Route path={PROJECT_PATHS}>
      Selected node: {selectedNode}
      {dag.nodes.map((n) => (
        <button key={n.name} onClick={() => navigateToNode(n)}>
          {n.name}
        </button>
      ))}
    </Route>
  );
};

export const setIdTokenForAccount = (account: Account) => {
  window.localStorage.setItem('id-token', generateIdTokenForAccount(account));
};

export const click: typeof userEvent.click = (...args) => {
  act(() => userEvent.click(...args));
};

export const hover: typeof userEvent.hover = (...args) => {
  act(() => userEvent.hover(...args));
};

export const paste: typeof userEvent.paste = (...args) => {
  act(() => userEvent.paste(...args));
};

export const tab: typeof userEvent.tab = (...args) => {
  act(() => userEvent.tab(...args));
};

// NOTE: The return type doesn't match userEvent.type,
// as our "type" command always returns a promise, whereas
// theirs only returns a promise if the input arguments contain a
// "delay". This has caused problems for us in the past,
// which is why we have a custom return type here.
export const type = async (
  element: TargetElement,
  text: string,
  userOpts?: ITypeOpts,
) => {
  await act(async () => {
    await userEvent.type(element, text, userOpts);
  });
};

export const getUrlState = (): UrlState => {
  const searchParams = new URLSearchParams(window.location.search);
  return JSON.parse(atob(searchParams.get('view') || '{}'));
};

export const generateTutorialView = (tutorialId: string) => {
  return btoa(JSON.stringify({tutorialId}));
};
