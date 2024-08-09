import {render, screen} from '@testing-library/react';
import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {click} from '@dash-frontend/testHelpers';
import {Tabs} from '@pachyderm/components';

describe('RouterTabs', () => {
  const TabsComponent = () => {
    return (
      <BrowserRouter>
        <Tabs.RouterTabs basePath="/:tabId?" basePathTabId={'one'}>
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
      </BrowserRouter>
    );
  };

  afterEach(() => {
    window.history.replaceState('', '', '/');
  });

  it('should display the basePathTabId at the basePath', () => {
    render(<TabsComponent />);

    const panel1 = screen.getByTestId('panel-one');
    expect(panel1).toBeVisible();
  });

  it('should display the correct tab based on the url', () => {
    window.history.replaceState('', '', '/two');
    render(<TabsComponent />);

    const panel2 = screen.getByTestId('panel-two');
    expect(panel2).toBeVisible();
  });

  it('should navigate the browser, and display the correct tab on selection', async () => {
    render(<TabsComponent />);

    const tab = screen.getByTestId('Tab__three');

    await click(tab);

    expect(screen.getByTestId('panel-three')).toBeVisible();
    expect(window.location.pathname).toBe('/three');
  });
});
