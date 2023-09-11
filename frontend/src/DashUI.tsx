import React, {Suspense, lazy} from 'react';
import {BrowserRouter, Redirect, Route, Switch} from 'react-router-dom';

import AuthenticatedRoute from '@dash-frontend/components/AuthenticatedRoute';
import LoadingSkeleton from '@dash-frontend/components/LoadingSkeleton';
import AnalyticsProvider from '@dash-frontend/providers/AnalyticsProvier';
import ApolloProvider from '@dash-frontend/providers/ApolloProvider';
import ErrorView from '@dash-frontend/views/ErrorView';
import {
  LINEAGE_PATH,
  PROJECT_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  NotificationBannerProvider,
  TutorialModalBodyProvider,
} from '@pachyderm/components';

import ErrorBoundaryProvider from './providers/ErrorBoundaryProvider';
import LoggedInProvider from './providers/LoggedInProvider';

const Landing = lazy(() => import('@dash-frontend/views/Landing/Landing'));
const Project = lazy(() => import('@dash-frontend/views/Project/Project'));

const DashUI: React.FC = () => {
  return (
    <BrowserRouter>
      <LoggedInProvider>
        <ApolloProvider>
          <ErrorBoundaryProvider>
            <AnalyticsProvider>
              <TutorialModalBodyProvider>
                <NotificationBannerProvider>
                  <main id="main">
                    <Suspense fallback={<LoadingSkeleton />}>
                      <Switch>
                        <Route
                          path="/"
                          exact
                          component={AuthenticatedRoute(Landing)}
                        />

                        <Route
                          path={[PROJECT_PATH, LINEAGE_PATH]}
                          component={AuthenticatedRoute(Project)}
                        />

                        <Route path="/not-found" exact component={ErrorView} />
                        <Redirect to={'/not-found'} />
                      </Switch>
                    </Suspense>
                  </main>
                </NotificationBannerProvider>
              </TutorialModalBodyProvider>
            </AnalyticsProvider>
          </ErrorBoundaryProvider>
        </ApolloProvider>
      </LoggedInProvider>
    </BrowserRouter>
  );
};

export default DashUI;
