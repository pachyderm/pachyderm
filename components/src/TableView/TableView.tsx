import noop from 'lodash/noop';
import React from 'react';

import {ErrorRetry} from 'ErrorRetry';
import {Page} from 'Page';

import {TableViewBody} from './components/TableViewBody';
import {TableViewHeader} from './components/TableViewHeader';

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

export default Object.assign(TableView, {
  Body: TableViewBody,
  Header: TableViewHeader,
});
