/* eslint-disable @typescript-eslint/no-explicit-any */
import {SUBSCRIPTION_INTERVAL} from '@dash-backend/constants/subscription';
import {generateIdTokenForAccount} from '@dash-backend/testHelpers';
import {Account, NodeType} from '@graphqlTypes';
import {act} from '@testing-library/react';
import userEvent, {ITypeOpts, TargetElement} from '@testing-library/user-event';
import React, {ReactElement} from 'react';
import {BrowserRouter, Route} from 'react-router-dom';

import ApolloProvider from '@dash-frontend/providers/ApolloProvider';

import {useProjectDagsData} from './hooks/useProjectDAGsData';
import useRouteController from './hooks/useRouteController';
import {UrlState} from './hooks/useUrlQueryState';
import useUrlState from './hooks/useUrlState';
import {DagDirection} from './lib/types';
import LoggedInProvider from './providers/LoggedInProvider';
import useDeletePipelineButton from './views/Project/components/ProjectSidebar/components/DeletePipelineButton/hooks/useDeletePipelineButton';
import {LINEAGE_PATH} from './views/Project/constants/projectPaths';

export {default as server} from '@dash-backend/index';
export {default as mockServer} from '@dash-backend/mock';

// Two timeout intervals should ensure that new data
// is hydrated if render is performed between events
export const SUBSCRIPTION_TIMEOUT = SUBSCRIPTION_INTERVAL * 2;

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

export const MockDag = () => {
  const {projectId, jobId} = useUrlState();
  const {selectedNode, navigateToNode} = useRouteController();
  const {dags, loading} = useProjectDagsData({
    projectId,
    nodeHeight: 60,
    nodeWidth: 120,
    direction: DagDirection.RIGHT,
    jobSetId: jobId,
  });
  const {onDelete, updating} = useDeletePipelineButton();

  if (loading || updating) return <span>Loading</span>;

  return (
    <Route path={LINEAGE_PATH}>
      <div>Selected node: {selectedNode}</div>
      {(dags || []).map((dag, i) => {
        return (
          <div key={i}>
            <div>
              {i} id: {dag.id}
            </div>
            {dag.nodes.map((node, i) => {
              return (
                <div key={node.id}>
                  <button onClick={() => navigateToNode(node)}>
                    {i} node id: {node.id}
                  </button>
                  <span>
                    {i} node name: {node.name}
                  </span>
                  <span>
                    {i} node type: {node.type}
                  </span>
                  <span>
                    {i} node state: {node.state}
                  </span>
                  {node.type === NodeType.PIPELINE && (
                    <button onClick={onDelete}>delete {node.id}</button>
                  )}
                </div>
              );
            })}
            {dag.links.map((link, i) => {
              return (
                <div key={link.id}>
                  <span>
                    {i} link source: {link.source}
                  </span>
                  <span>
                    {i} link target: {link.target}
                  </span>
                  <span>
                    {i} link state: {link.state}
                  </span>
                </div>
              );
            })}
          </div>
        );
      })}
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
