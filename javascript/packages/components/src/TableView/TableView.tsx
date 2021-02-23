import noop from 'lodash/noop';
import React from 'react';

import ErrorRetry from 'ErrorRetry';
import Page from 'Page';

export interface TableViewProps {
  title: string;
  error?: boolean;
  errorMessage: string;
  retry?: () => void;
  hasDrawerPadding?: boolean;
}

const TableView: React.FC<TableViewProps> = ({
  error,
  errorMessage,
  retry = noop,
  title,
  children,
  hasDrawerPadding = false,
}) => {
  return (
    <Page fullHeight hasDrawerPadding={hasDrawerPadding} title={title}>
      {error ? <ErrorRetry retry={retry}>{errorMessage}</ErrorRetry> : children}
    </Page>
  );
};

export default TableView;
