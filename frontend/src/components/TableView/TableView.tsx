import capitalize from 'lodash/capitalize';
import React from 'react';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import {
  Tabs,
  Button,
  FilterSVG,
  ArrowRightSVG,
  Icon,
} from '@pachyderm/components';

import styles from './TableView.module.css';

type TableViewProps = {
  children?: React.ReactNode;
  title: string;
  noun: string;
  tabsBasePath: string;
  tabs: Record<string, string>;
  selectedItems: string[];
  singleRowSelection?: boolean;
  filtersExpanded: boolean;
  setFiltersExpanded: React.Dispatch<React.SetStateAction<boolean>>;
};

const TableView: React.FC<TableViewProps> = ({
  title,
  noun,
  tabsBasePath,
  tabs,
  selectedItems,
  singleRowSelection = false,
  filtersExpanded,
  setFiltersExpanded,
  children,
}) => {
  const firstTabKey = Object.keys(tabs)[0];
  const firstTabValue = Object.values(tabs)[0];

  return (
    <div className={styles.wrapper}>
      <BrandedTitle title={title} />
      <h4 className={styles.pageTitle}>{title}</h4>
      <div className={styles.base}>
        <Tabs.RouterTabs basePathTabId={firstTabKey} basePath={tabsBasePath}>
          <Tabs.TabsHeader className={styles.tabsHeader}>
            <Tabs.Tab id={firstTabKey}>{capitalize(firstTabValue)}</Tabs.Tab>
            <span className={styles.filterInstructions}>
              {selectedItems.length === 0 && (
                <>
                  {singleRowSelection
                    ? 'Select a row to view detailed info '
                    : 'Select rows to view detailed info '}
                  <Icon smaller>
                    <ArrowRightSVG />
                  </Icon>
                </>
              )}
              {selectedItems.length > 0 && (
                <>
                  Detailed info for{' '}
                  {singleRowSelection
                    ? selectedItems[0]
                    : `${selectedItems.length} ${noun}`}
                  {selectedItems.length > 1 ? 's' : ''}
                  <Icon smaller>
                    <ArrowRightSVG />
                  </Icon>
                </>
              )}
            </span>
            {Object.entries(tabs)
              .slice(1)
              .map(([key, value]) => (
                <Tabs.Tab
                  key={key}
                  id={key}
                  disabled={selectedItems.length === 0}
                >
                  {capitalize(value)}
                </Tabs.Tab>
              ))}
            <span className={styles.filterButton}>
              <Button
                IconSVG={FilterSVG}
                onClick={() => setFiltersExpanded(!filtersExpanded)}
                buttonType="ghost"
                aria-label="expand filters"
              />
            </span>
          </Tabs.TabsHeader>
          {children}
        </Tabs.RouterTabs>
      </div>
    </div>
  );
};

export default TableView;
