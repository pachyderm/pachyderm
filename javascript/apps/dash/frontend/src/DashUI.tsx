import React from 'react';
import {BrowserRouter, Route, Switch} from 'react-router-dom';

import AuthenticatedRoute from '@dash-frontend/components/AuthenticatedRoute';
import Header from '@dash-frontend/components/Header';
import ApolloProvider from '@dash-frontend/providers/ApolloProvider';
import Landing from '@dash-frontend/views/Landing';
import Project from '@dash-frontend/views/Project';

const DashUI: React.FC = () => {
  return (
    <BrowserRouter>
      <ApolloProvider>
        <Header />
        <main id="main">
          <Switch>
            <Route path="/" exact component={AuthenticatedRoute(Landing)} />

            <Route
              path="/project/:projectId"
              component={AuthenticatedRoute(Project)}
            />

            {/* TODO: Error page(s) */}
            <Route path="/error" exact component={() => <>Error</>} />
          </Switch>
        </main>
      </ApolloProvider>
    </BrowserRouter>
  );
};

export default DashUI;
