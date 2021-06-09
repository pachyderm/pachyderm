import {TableView} from '@pachyderm/components';
import React from 'react';

import Sidebar from '@dash-frontend/components/Sidebar';
import View from '@dash-frontend/components/View';

import LandingHeader from '../LandingHeader';

import styles from './LandingSkeleton.module.css';

const LandingSkeleton: React.FC = () => {
  return (
    <>
      <LandingHeader />
      <div className={styles.base}>
        <View>
          <TableView title="Projects" errorMessage="Error loading projects">
            <TableView.Header heading="Projects" headerButtonHidden />
            <TableView.Body initialActiveTabId={'All'} showSkeleton={false}>
              <TableView.Body.Header>
                <TableView.Body.Tabs placeholder="">
                  <TableView.Body.Tabs.Tab id="All" count={0}>
                    All
                  </TableView.Body.Tabs.Tab>
                </TableView.Body.Tabs>
              </TableView.Body.Header>
              <TableView.Body.Content id={'All'}>
                <table className={styles.table}>
                  <tbody>
                    <tr className={styles.loadingProject}>
                      <td className={styles.loadingProjectCell} />
                      <td className={styles.loadingProjectCell} />
                      <td className={styles.loadingProjectCell} />
                      <td className={styles.loadingProjectCell} />
                      <td className={styles.loadingProjectCell} />
                    </tr>
                  </tbody>
                </table>
              </TableView.Body.Content>
            </TableView.Body>
          </TableView>
        </View>
        <Sidebar />
      </div>
    </>
  );
};

export default LandingSkeleton;
