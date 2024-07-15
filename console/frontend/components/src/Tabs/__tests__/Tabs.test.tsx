import {fireEvent, render, screen} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import {Tabs} from '../';

describe('Tabs', () => {
  const renderTestbed = (renderWhenHidden?: boolean) => {
    render(
      <Tabs initialActiveTabId={'one'}>
        <Tabs.TabsHeader>
          <Tabs.Tab id="one">One</Tabs.Tab>
          <Tabs.Tab id="two">Two</Tabs.Tab>
          <Tabs.Tab id="three">Three</Tabs.Tab>
        </Tabs.TabsHeader>

        <Tabs.TabPanel
          id="one"
          data-testid={'panel-one'}
          renderWhenHidden={renderWhenHidden}
        >
          One Content
        </Tabs.TabPanel>
        <Tabs.TabPanel
          id="two"
          data-testid={'panel-two'}
          renderWhenHidden={renderWhenHidden}
        >
          Two Content
        </Tabs.TabPanel>
        <Tabs.TabPanel
          id="three"
          data-testid={'panel-three'}
          renderWhenHidden={renderWhenHidden}
        >
          Three Content
        </Tabs.TabPanel>
      </Tabs>,
    );
  };

  const assertTab = (id: string) => {
    const tabPanel = screen.queryByTestId(`panel-${id}`);

    return {
      toBeShown: () => expect(tabPanel).toBeVisible(),
      toBeHidden: () => expect(tabPanel).not.toBeVisible(),
      notToBeRendered: () => expect(tabPanel).not.toBeInTheDocument(),
    };
  };

  const activateTab = async (id: string) => {
    const tab = screen.getByTestId(`Tab__${id}`);

    await click(tab);
  };

  /* eslint-disable-next-line jest/expect-expect */
  it('should initially display the correct active tab', () => {
    renderTestbed();

    assertTab('one').toBeShown();
    assertTab('two').toBeHidden();
    assertTab('three').toBeHidden();
  });

  /* eslint-disable-next-line jest/expect-expect */
  it('should be able to navigate between tabs', async () => {
    renderTestbed();

    await activateTab('two');

    assertTab('two').toBeShown();
    assertTab('one').toBeHidden();
    assertTab('three').toBeHidden();
  });

  /* eslint-disable-next-line jest/expect-expect */
  it('should be able to navigate between tabs with hidden tabs not rendered', async () => {
    renderTestbed(false);

    await activateTab('two');

    assertTab('two').toBeShown();
    assertTab('one').notToBeRendered();
    assertTab('three').notToBeRendered();
  });

  describe('keyboard behavior', () => {
    it('should allow users to navigate between tabs using the right arrow key', () => {
      renderTestbed();

      const tab1 = screen.getByTestId('Tab__one');
      const panel1 = screen.getByTestId('panel-one');

      const tab2 = screen.getByTestId('Tab__two');
      const panel2 = screen.getByTestId('panel-two');

      const tab3 = screen.getByTestId('Tab__three');
      const panel3 = screen.getByTestId('panel-three');

      tab1.focus();

      fireEvent.keyDown(tab1, {key: 'ArrowRight'});

      expect(tab2).toHaveFocus();
      expect(panel2).toBeVisible();

      fireEvent.keyDown(tab2, {key: 'ArrowRight'});

      expect(tab3).toHaveFocus();
      expect(panel3).toBeVisible();

      fireEvent.keyDown(tab3, {key: 'ArrowRight'});

      expect(tab1).toHaveFocus();
      expect(panel1).toBeVisible();
    });

    it('should allow users to navigate between tabs using the left arrow key', () => {
      renderTestbed();

      const tab1 = screen.getByTestId('Tab__one');
      const panel1 = screen.getByTestId('panel-one');

      const tab2 = screen.getByTestId('Tab__two');
      const panel2 = screen.getByTestId('panel-two');

      const tab3 = screen.getByTestId('Tab__three');
      const panel3 = screen.getByTestId('panel-three');

      tab1.focus();

      fireEvent.keyDown(tab1, {key: 'ArrowLeft'});

      expect(tab3).toHaveFocus();
      expect(panel3).toBeVisible();

      fireEvent.keyDown(tab3, {key: 'ArrowLeft'});

      expect(tab2).toHaveFocus();
      expect(panel2).toBeVisible();

      fireEvent.keyDown(tab2, {key: 'ArrowLeft'});

      expect(tab1).toHaveFocus();
      expect(panel1).toBeVisible();
    });

    it('should allow users to navigate to the active panel using the down arrow key', async () => {
      renderTestbed();

      const tab1 = screen.getByTestId('Tab__one');
      const panel1 = screen.getByTestId('panel-one');

      await click(tab1);

      fireEvent.keyDown(tab1, {key: 'ArrowDown'});

      expect(panel1).toHaveFocus();
    });
  });
});
