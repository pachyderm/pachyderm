import React from 'react';

import {Tabs} from '@pachyderm/components';
import {TabProps} from '@pachyderm/components/Tabs/components/Tab/Tab';

import styles from './BodyHeaderTab.module.css';

export interface BodyHeaderTabProps extends TabProps {
  count: number | string;
}

const BodyHeaderTab: React.FC<BodyHeaderTabProps> = ({
  id,
  // count,
  children,
  ...rest
}) => {
  return (
    <Tabs.Tab id={id} {...rest}>
      <div className={styles.container}>
        {children}
        {/* FRON: https://pachyderm.atlassian.net/browse/FRON-1318 */}
        {/* <div className={styles.count}>{count}</div> */}
      </div>
    </Tabs.Tab>
  );
};

export default BodyHeaderTab;
