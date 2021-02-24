import {fireEvent, render} from '@testing-library/react';
import React from 'react';

import {click} from 'testHelpers';

import {Tabs} from '../';

describe('Tabs', () => {
  const renderTestBed = () => {
    const testUtils = render(
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
      </Tabs>,
    );

    const assertTab = (id: string) => {
      const tabPanel = testUtils.getByTestId(`panel-${id}`);

      return {
        toBeShown: () => expect(tabPanel).toBeVisible(),
        toBeHidden: () => expect(tabPanel).not.toBeVisible(),
      };
    };

    const activateTab = (id: string) => {
      const tab = testUtils.getByTestId(`Tab__${id}`);

      click(tab);
    };

    return {
      assertTab,
      activateTab,
      ...testUtils,
    };
  };

  it('should initially display the correct active tab', () => {
    const {assertTab} = renderTestBed();

    assertTab('one').toBeShown();
    assertTab('two').toBeHidden();
    assertTab('three').toBeHidden();
  });

  it('should be able to navigate between tabs', () => {
    const {activateTab, assertTab} = renderTestBed();

    activateTab('two');

    assertTab('two').toBeShown();
    assertTab('one').toBeHidden();
    assertTab('three').toBeHidden();
  });

  describe('keyboard behavior', () => {
    it('should allow users to navigate between tabs using the right arrow key', () => {
      const {getByTestId} = renderTestBed();

      const tab1 = getByTestId('Tab__one');
      const panel1 = getByTestId('panel-one');

      const tab2 = getByTestId('Tab__two');
      const panel2 = getByTestId('panel-two');

      const tab3 = getByTestId('Tab__three');
      const panel3 = getByTestId('panel-three');

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
      const {getByTestId} = renderTestBed();

      const tab1 = getByTestId('Tab__one');
      const panel1 = getByTestId('panel-one');

      const tab2 = getByTestId('Tab__two');
      const panel2 = getByTestId('panel-two');

      const tab3 = getByTestId('Tab__three');
      const panel3 = getByTestId('panel-three');

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

    it('should allow users to navigate to the active panel using the down arrow key', () => {
      const {getByTestId} = renderTestBed();

      const tab1 = getByTestId('Tab__one');
      const panel1 = getByTestId('panel-one');

      click(tab1);

      fireEvent.keyDown(tab1, {key: 'ArrowDown'});

      expect(panel1).toHaveFocus();
    });
  });
});
