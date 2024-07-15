import classnames from 'classnames';
import React from 'react';

import {
  DefaultDropdown,
  Group,
  ButtonGroup,
  ChevronLeftSVG,
  ChevronRightSVG,
  Button,
} from '@pachyderm/components';

import styles from './Pager.module.css';

type PagerProps = {
  page: number;
  pageCount?: number;
  nextPageDisabled?: boolean;
  updatePage: (page: number) => void;
  pageSizes: number[];
  updatePageSize: React.Dispatch<React.SetStateAction<number>>;
  pageSize: number;
  elementName?: string;
  hasTopBorder?: boolean;
  highlightFirstPageNavigation?: boolean;
};

interface SimplePagerProps
  extends Omit<PagerProps, 'pageSizes' | 'updatePageSize'> {
  contentLength: number;
}

export const Pager: React.FC<PagerProps> = ({
  page,
  pageCount,
  nextPageDisabled = false,
  updatePage,
  pageSizes,
  updatePageSize,
  pageSize,
  elementName,
  hasTopBorder,
  highlightFirstPageNavigation,
}) => {
  const secondPageAvailable =
    highlightFirstPageNavigation && !nextPageDisabled && page === 1;
  return (
    <div
      className={classnames(styles.base, {[styles.hasTopBorder]: hasTopBorder})}
      data-testid="Pager__pager"
    >
      <Group className={styles.pages}>
        <DefaultDropdown
          items={pageSizes.map((i) => ({
            id: i.toString(),
            content: i.toString(),
            closeOnClick: true,
          }))}
          openUpwards
          buttonOpts={{buttonType: 'quaternary'}}
          storeSelected
          initialSelectId={pageSize.toString()}
          onSelect={(val) => {
            updatePageSize(parseInt(val));
            updatePage(1);
          }}
        >
          {pageSize}
        </DefaultDropdown>
        <span className={styles.pageDropdown}>
          {elementName || 'item'}s per page
        </span>
      </Group>
      <div className={styles.navigation}>
        {page && (
          <span className={styles.pageCount}>
            Page {page}
            {pageCount ? ` of ${pageCount}` : ''}
          </span>
        )}
        <ButtonGroup className={styles.navigationButtons}>
          <Button
            IconSVG={ChevronLeftSVG}
            data-testid="Pager__backward"
            buttonType="ghost"
            onClick={() => updatePage(page - 1)}
            disabled={page === 1}
          />
          <Button
            IconSVG={ChevronRightSVG}
            data-testid="Pager__forward"
            buttonType={secondPageAvailable ? 'secondary' : 'ghost'}
            onClick={() => updatePage(page + 1)}
            disabled={
              nextPageDisabled || (pageCount ? page >= pageCount : false)
            }
          >
            {secondPageAvailable ? 'Next Page' : ''}
          </Button>
        </ButtonGroup>
      </div>
    </div>
  );
};

export const SimplePager: React.FC<SimplePagerProps> = ({
  page,
  pageCount,
  nextPageDisabled = false,
  updatePage,
  pageSize,
  elementName,
  contentLength,
  hasTopBorder,
  highlightFirstPageNavigation,
}) => {
  const secondPageAvailable =
    highlightFirstPageNavigation && !nextPageDisabled && page === 1;
  return (
    <div
      className={classnames(styles.base, {[styles.hasTopBorder]: hasTopBorder})}
      data-testid="Pager__simplePager"
    >
      <div className={styles.navigation}>
        {page && contentLength !== 0 && (
          <span>
            {elementName || 'Item'}s {(page - 1) * pageSize + 1} -{' '}
            {Math.min(contentLength, page * pageSize)}{' '}
            {pageCount ? `of  ${contentLength}` : ''}
          </span>
        )}
      </div>
      <ButtonGroup className={styles.navigationButtons}>
        <Button
          IconSVG={ChevronLeftSVG}
          data-testid="Pager__backward"
          buttonType="ghost"
          onClick={() => updatePage(page - 1)}
          disabled={page === 1}
        />
        <Button
          IconSVG={ChevronRightSVG}
          data-testid="Pager__forward"
          buttonType={secondPageAvailable ? 'secondary' : 'ghost'}
          onClick={() => updatePage(page + 1)}
          disabled={nextPageDisabled || (pageCount ? page >= pageCount : false)}
        >
          {secondPageAvailable ? 'Next Page' : ''}
        </Button>
      </ButtonGroup>
    </div>
  );
};
