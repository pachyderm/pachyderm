import React, {Suspense, lazy} from 'react';
import {BrowserRouter, Redirect, Route, Switch} from 'react-router-dom';

import AuthenticatedRoute from '@dash-frontend/components/AuthenticatedRoute';
import LoadingSkeleton from '@dash-frontend/components/LoadingSkeleton';
import ApolloProvider from '@dash-frontend/providers/ApolloProvider';
import ErrorView from '@dash-frontend/views/ErrorView';
const Landing = lazy(
  () => import(/* webpackPrefetch: true */ '@dash-frontend/views/Landing'),
);
const Project = lazy(
  () => import(/* webpackPrefetch: true */ '@dash-frontend/views/Project'),
);

const DashUI: React.FC = () => {
  return (
    <BrowserRouter>
      <ApolloProvider>
        <main id="main">
          <Suspense fallback={<LoadingSkeleton />}>
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
          </Suspense>
        </main>
      </ApolloProvider>
    </BrowserRouter>
  );
};

export default DashUI;
