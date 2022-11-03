import classnames from 'classnames';
import noop from 'lodash/noop';
import React from 'react';

import {Group} from 'Group';
import {Page} from 'Page';

import {Button, ButtonProps} from '../../../Button';

import styles from './TableViewHeader.module.css';

type TableViewHeaderProps = {
  heading?: string;
  headerButtonText?: string;
  headerButtonAction?: () => void;
  headerButtonDisabled?: boolean;
  headerButtonHidden?: boolean;
};

export const TableViewHeaderButton: React.FC<ButtonProps> = ({
  children,
  className,
  ...rest
}) => (
  <Button
    className={classnames(styles.button, className)}
    data-testid="TableViewHeader__button"
    {...rest}
  >
    {children}
  </Button>
);

const TableViewHeader: React.FC<TableViewHeaderProps> = ({
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
          <Page.Heading>{heading}</Page.Heading>
          {!headerButtonHidden && (
            <TableViewHeaderButton
              onClick={headerButtonAction}
              className={styles.button}
              disabled={headerButtonDisabled}
              data-testid="TableViewHeader__button"
            >
              {headerButtonText}
            </TableViewHeaderButton>
          )}
        </>
      )}
    </Group>
  );
};

export default Object.assign(TableViewHeader, {
  Button: TableViewHeaderButton,
  Heading: Page.Heading,
});
