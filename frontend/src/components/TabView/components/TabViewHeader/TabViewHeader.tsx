import classnames from 'classnames';
import noop from 'lodash/noop';
import React from 'react';

import {Group} from '@pachyderm/components';
import {Button, ButtonProps} from '@pachyderm/components/Button';

import styles from './TabViewHeader.module.css';

type TabViewHeaderProps = {
  children?: React.ReactNode;
  heading?: string;
  headerButtonText?: string;
  headerButtonAction?: () => void;
  headerButtonDisabled?: boolean;
  headerButtonHidden?: boolean;
};

export const TabViewHeaderButton: React.FC<ButtonProps> = ({
  children,
  className,
  ...rest
}) => (
  <Button
    className={classnames(styles.button, className)}
    data-testid="TabViewHeader__button"
    {...rest}
  >
    {children}
  </Button>
);

const TabViewHeader: React.FC<TabViewHeaderProps> = ({
  heading = '',
  headerButtonText = '',
  headerButtonAction = noop,
  headerButtonDisabled = false,
  headerButtonHidden = false,
  children,
}) => {
  return (
    <Group justify="stretch">
      {children ? (
        children
      ) : (
        <>
          <h2 className={styles.heading}>{heading}</h2>
          {!headerButtonHidden && (
            <TabViewHeaderButton
              onClick={headerButtonAction}
              className={styles.button}
              disabled={headerButtonDisabled}
              data-testid="TabViewHeader__button"
            >
              {headerButtonText}
            </TabViewHeaderButton>
          )}
        </>
      )}
    </Group>
  );
};

export default Object.assign(TabViewHeader, {
  Button: TabViewHeaderButton,
});
