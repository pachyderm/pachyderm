import React, {SelectHTMLAttributes} from 'react';

import useTabPanel from 'Tabs/hooks/useTabPanel';

import styles from './TabPanel.module.css';

export interface TabPanelProps extends SelectHTMLAttributes<HTMLSelectElement> {
  id: string;
}

const TabPanel: React.FC<TabPanelProps> = ({children, id, ...rest}) => {
  const {isShown} = useTabPanel(id);

  return (
    <section
      {...rest}
      role="tabpanel"
      hidden={!isShown}
      aria-labelledby={id}
      className={styles.base}
      tabIndex={-1}
    >
      {children}
    </section>
  );
};

export default TabPanel;
