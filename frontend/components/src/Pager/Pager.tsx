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
  updatePage: React.Dispatch<React.SetStateAction<number>>;
  pageSizes: number[];
  updatePageSize: React.Dispatch<React.SetStateAction<number>>;
  pageSize: number;
  elementName?: string;
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
}) => {
  return (
    <div className={styles.base}>
      <Group className={styles.pages}>
        <DefaultDropdown
          items={pageSizes.map((i) => ({
            id: i.toString(),
            content: i.toString(),
            closeOnClick: true,
          }))}
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
            buttonType="ghost"
            onClick={() => updatePage(page + 1)}
            disabled={
              nextPageDisabled || (pageCount ? page >= pageCount : false)
            }
          />
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
}) => {
  return (
    <div className={styles.base}>
      <div className={styles.navigation}>
        {page && pageCount && (
          <span>
            {elementName || 'Item'}s {(page - 1) * pageSize + 1} -{' '}
            {Math.min(contentLength, page * pageSize)} of {contentLength}
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
          buttonType="ghost"
          onClick={() => updatePage(page + 1)}
          disabled={nextPageDisabled || (pageCount ? page >= pageCount : false)}
        />
      </ButtonGroup>
    </div>
  );
};
