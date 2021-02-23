import React from 'react';
import {BrowserRouter, Route, Switch} from 'react-router-dom';

import AuthenticatedRoute from 'components/AuthenticatedRoute';
import ApolloProvider from 'providers/ApolloProvider';
import Home from 'views/Home';

const DashUI: React.FC = () => {
  return (
    <BrowserRouter>
      <ApolloProvider>
        <main id="main">
          <Switch>
            <Route path="/" exact component={AuthenticatedRoute(Home)} />

            {/* TODO: Error page(s) */}
            <Route path="/error" exact component={() => <>Error</>} />
          </Switch>
        </main>
      </ApolloProvider>
    </BrowserRouter>
  );
};

export default DashUI;
