import noop from 'lodash/noop';
import React from 'react';

import {ErrorRetry} from '@dash-frontend/components/ErrorRetry';
import {Page} from '@pachyderm/components';

import {TabViewBody} from './components/TabViewBody';
import {TabViewHeader} from './components/TabViewHeader';

export interface TabViewProps {
  children?: React.ReactNode;
  error?: boolean;
  errorMessage: string;
  retry?: () => void;
  hasDrawerPadding?: boolean;
}

const TabView: React.FC<TabViewProps> = ({
  error,
  errorMessage,
  retry = noop,
  children,
  hasDrawerPadding = false,
}) => {
  return (
    <Page fullHeight hasDrawerPadding={hasDrawerPadding}>
      {error ? <ErrorRetry retry={retry}>{errorMessage}</ErrorRetry> : children}
    </Page>
  );
};

export default Object.assign(TabView, {
  Body: TabViewBody,
  Header: TabViewHeader,
});
