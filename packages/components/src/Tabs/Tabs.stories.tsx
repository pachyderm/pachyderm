import React from 'react';
import {MemoryRouter} from 'react-router-dom';

import {Tabs} from './';

export default {title: 'Tabs'};

export const Default = () => {
  return (
    <Tabs initialActiveTabId={'one'}>
      <Tabs.TabsHeader>
        <Tabs.Tab id="one">One</Tabs.Tab>
        <Tabs.Tab id="two">Two</Tabs.Tab>
        <Tabs.Tab id="three">Three</Tabs.Tab>
      </Tabs.TabsHeader>

      <Tabs.TabPanel id="one" data-testid={'panel-one'}>
        One Content
      </Tabs.TabPanel>
      <Tabs.TabPanel id="two" data-testid={'panel-two'}>
        Two Content
      </Tabs.TabPanel>
      <Tabs.TabPanel id="three" data-testid={'panel-three'}>
        Three Content
      </Tabs.TabPanel>
    </Tabs>
  );
};

export const RouterTabs = () => {
  return (
    <MemoryRouter initialEntries={['/one']}>
      {/*
          Storybook displays each canvas in an iframe, and uses a "path" parameter
          to determine which story to display. Therefore, it doesn't play well with the
          history API or BrowserRouter
      */}
      <Tabs.RouterTabs basePath="/:tabId?" basePathTabId="one">
        <Tabs.TabsHeader>
          <Tabs.Tab id="one">One</Tabs.Tab>
          <Tabs.Tab id="two">Two</Tabs.Tab>
          <Tabs.Tab id="three">Three</Tabs.Tab>
        </Tabs.TabsHeader>

        <Tabs.TabPanel id="one" data-testid={'panel-one'}>
          One Content
        </Tabs.TabPanel>
        <Tabs.TabPanel id="two" data-testid={'panel-two'}>
          Two Content
        </Tabs.TabPanel>
        <Tabs.TabPanel id="three" data-testid={'panel-three'}>
          Three Content
        </Tabs.TabPanel>
      </Tabs.RouterTabs>
    </MemoryRouter>
  );
};
