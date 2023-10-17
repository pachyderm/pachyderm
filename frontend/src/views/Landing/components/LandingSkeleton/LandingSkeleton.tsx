import React from 'react';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import Sidebar from '@dash-frontend/components/Sidebar';
import {TabView} from '@dash-frontend/components/TabView';
import View from '@dash-frontend/components/View';

import LandingHeader from '../LandingHeader';

import styles from './LandingSkeleton.module.css';

const LandingSkeleton: React.FC = () => {
  return (
    <>
      <BrandedTitle title="Projects" />
      <LandingHeader />
      <div className={styles.base}>
        <View>
          <TabView errorMessage="Error loading projects">
            <TabView.Header heading="Projects" />
            <TabView.Body initialActiveTabId={'All'} showSkeleton={false}>
              <TabView.Body.Header>
                <TabView.Body.Tabs placeholder="">
                  <TabView.Body.Tabs.Tab id="All" count={0}>
                    All
                  </TabView.Body.Tabs.Tab>
                </TabView.Body.Tabs>
              </TabView.Body.Header>
              <TabView.Body.Content id={'All'}>
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
              </TabView.Body.Content>
            </TabView.Body>
          </TabView>
        </View>
        <Sidebar className={styles.sidebar} />
      </div>
    </>
  );
};

export default LandingSkeleton;
