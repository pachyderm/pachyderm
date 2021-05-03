import React from 'react';
import {BrowserRouter, Redirect, Route, Switch} from 'react-router-dom';

import AuthenticatedRoute from '@dash-frontend/components/AuthenticatedRoute';
import ApolloProvider from '@dash-frontend/providers/ApolloProvider';
import ErrorView from '@dash-frontend/views/ErrorView';
import Landing from '@dash-frontend/views/Landing';
import Project from '@dash-frontend/views/Project';

const DashUI: React.FC = () => {
  return (
    <BrowserRouter>
      <ApolloProvider>
        <main id="main">
          <Switch>
            <Route path="/" exact component={AuthenticatedRoute(Landing)} />

            <Route
              path="/project/:projectId"
              component={AuthenticatedRoute(Project)}
            />

            <Route path="/not-found" exact component={ErrorView} />
            <Route path="/unauthenticated" exact component={ErrorView} />
            <Route path="/error" exact component={ErrorView} />

            <Redirect to={'/not-found'} />
          </Switch>
        </main>
      </ApolloProvider>
    </BrowserRouter>
  );
};

export default DashUI;
