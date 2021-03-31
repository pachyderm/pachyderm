import React from 'react';
import {Route, Switch} from 'react-router';

import LandingHeader from './components/LandingHeader';
import ProjectHeader from './components/ProjectHeader';
import styles from './Header.module.css';

const Header: React.FC = () => {
  return (
    <header className={styles.base}>
      <Switch>
        <Route path="/project/:projectId" component={ProjectHeader} />
        <Route path="*" component={LandingHeader} />
      </Switch>
    </header>
  );
};

export default Header;
