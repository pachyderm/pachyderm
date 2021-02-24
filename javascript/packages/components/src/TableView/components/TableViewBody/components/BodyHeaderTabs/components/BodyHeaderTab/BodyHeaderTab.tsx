import React from 'react';

import {Tabs} from 'Tabs';
import {TabProps} from 'Tabs/components/Tab/Tab';

import styles from './BodyHeaderTab.module.css';

export interface BodyHeaderTabProps extends TabProps {
  count: number;
}

const BodyHeaderTab: React.FC<BodyHeaderTabProps> = ({
  id,
  count,
  children,
  ...rest
}) => {
  return (
    <Tabs.Tab id={id} {...rest}>
      <div className={styles.container}>
        {children}
        <div className={styles.count}>{count}</div>
      </div>
    </Tabs.Tab>
  );
};

export default BodyHeaderTab;
