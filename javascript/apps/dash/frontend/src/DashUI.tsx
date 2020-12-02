import React from 'react';
import {BrowserRouter, Route, Switch} from 'react-router-dom';

import ApolloProvider from 'providers/ApolloProvider';
import HelloDash from 'views/HelloDash';

const DashUI: React.FC = () => {
  return (
    <BrowserRouter>
      <ApolloProvider>
        <main id="main">
          <Switch>
            <Route path="/" exact component={HelloDash} />
          </Switch>
        </main>
      </ApolloProvider>
    </BrowserRouter>
  );
};

export default DashUI;
