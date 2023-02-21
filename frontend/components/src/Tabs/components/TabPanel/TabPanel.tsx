import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import useTabPanel from '@pachyderm/components/Tabs/hooks/useTabPanel';

import styles from './TabPanel.module.css';

export interface TabPanelProps extends HTMLAttributes<HTMLDivElement> {
  id: string;
  className?: string;
  renderWhenHidden?: boolean;
}

const TabPanel: React.FC<TabPanelProps> = ({
  children,
  id,
  className,
  renderWhenHidden = true,
  ...rest
}) => {
  const {isShown} = useTabPanel(id);

  if (!renderWhenHidden && !isShown) {
    return null;
  }

  return (
    <div
      {...rest}
      role="tabpanel"
      hidden={!isShown}
      aria-labelledby={id}
      className={classnames(styles.base, className)}
      tabIndex={-1}
    >
      {children}
    </div>
  );
};

export default TabPanel;
