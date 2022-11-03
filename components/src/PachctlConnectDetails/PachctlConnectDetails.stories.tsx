import React from 'react';

import styles from './PachctlConnectDetails.module.css';

import {Connect, Verify} from './';

export default {title: 'PachCtl Connect Details'};

export const Default = () => {
  return (
    <>
      <div className={styles.base}>
        <Connect name="workspace" pachdAddress="workspaceAddress" />
      </div>
      <div className={styles.base}>
        <Verify>
          <tr>
            <td>pachctl</td>
            <td>1.0.0</td>
          </tr>
          <tr>
            <td>pachd</td>
            <td>1.0.0</td>
          </tr>
        </Verify>
      </div>
    </>
  );
};
