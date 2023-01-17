import {render, screen} from '@testing-library/react';
import React, {useState} from 'react';

import {click} from '@dash-frontend/testHelpers';

import {Pager} from '../';

describe('Pager', () => {
  const PagerComponent: React.FC<{
    intialPage?: number;
    defaultPageSize?: number;
  }> = ({intialPage, defaultPageSize}) => {
    const [page, setPage] = useState(intialPage || 1);
    const [pageSize, setPageSize] = useState(defaultPageSize || 5);

    const content = Array.from(Array(103).keys());

    return (
      <>
        {content
          .slice((page - 1) * pageSize, page * pageSize)
          .map((i) => i + ' ')}
        <Pager
          page={page}
          pageCount={Math.ceil(content.length / pageSize)}
          updatePage={setPage}
          pageSizes={[5, 10, 20]}
          updatePageSize={setPageSize}
          pageSize={pageSize}
        />
      </>
    );
  };

  it('should display paged data', async () => {
    render(<PagerComponent intialPage={2} />);

    expect(screen.getByText('5 6 7 8 9')).toBeInTheDocument();
    expect(screen.getByText('Page 2 of 21')).toBeInTheDocument();
  });

  it('should navigate through pages within boundaries', async () => {
    render(<PagerComponent defaultPageSize={20} />);

    const forwards = screen.getByTestId('Pager__forward');
    const backwards = screen.getByTestId('Pager__backward');

    expect(
      screen.getByText('0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19'),
    ).toBeInTheDocument();
    expect(screen.getByText('Page 1 of 6')).toBeInTheDocument();
    expect(backwards).toBeDisabled();

    await click(forwards);
    await click(forwards);
    await click(forwards);
    await click(forwards);
    await click(forwards);

    expect(screen.getByText('100 101 102')).toBeInTheDocument();
    expect(screen.getByText('Page 6 of 6')).toBeInTheDocument();
    expect(forwards).toBeDisabled();
  });

  it('should allow users to change the page size', async () => {
    render(<PagerComponent />);

    expect(screen.getByText('0 1 2 3 4')).toBeInTheDocument();
    expect(screen.getByText('Page 1 of 21')).toBeInTheDocument();

    await click(screen.getAllByText('5')[0]);
    await click(screen.getByText('20'));

    expect(
      screen.getByText('0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19'),
    ).toBeInTheDocument();
    expect(screen.getByText('Page 1 of 6')).toBeInTheDocument();

    await click(screen.getByTestId('Pager__forward'));

    expect(
      screen.getByText(
        '20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39',
      ),
    ).toBeInTheDocument();
    expect(screen.getByText('Page 2 of 6')).toBeInTheDocument();
  });
});
