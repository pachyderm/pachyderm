import React from 'react';
import {BrowserRouter, Route, Switch} from 'react-router-dom';

import AuthenticatedRoute from 'components/AuthenticatedRoute';
import Header from 'components/Header';
import ApolloProvider from 'providers/ApolloProvider';
import Landing from 'views/Landing';
import Project from 'views/Project';

const DashUI: React.FC = () => {
  return (
    <BrowserRouter>
      <ApolloProvider>
        <main id="main">
          <Header />
          <Switch>
            <Route path="/" exact component={AuthenticatedRoute(Landing)} />

            <Route
              path="/project/:projectId"
              exact
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
