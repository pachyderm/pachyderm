import {render} from '@testing-library/react';
import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {Tabs} from 'Tabs';
import {click} from 'testHelpers';

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
    const {getByTestId} = render(<TabsComponent />);

    const panel1 = getByTestId('panel-one');
    expect(panel1).toBeVisible();
  });

  it('should display the correct tab based on the url', () => {
    window.history.replaceState('', '', '/two');
    const {getByTestId} = render(<TabsComponent />);

    const panel2 = getByTestId('panel-two');
    expect(panel2).toBeVisible();
  });

  it('should navigate the browser, and display the correct tab on selection', async () => {
    const {getByTestId} = render(<TabsComponent />);

    const tab = getByTestId('Tab__three');

    click(tab);

    expect(getByTestId('panel-three')).toBeVisible();
    expect(window.location.pathname).toBe('/three');
  });
});
