import {screen, render} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import useCSVPreview from '../useCSVPreview';

describe('CSVPreview/hooks/useCSVPreview', () => {
  const server = setupServer();
  const Component = withContextProviders(() => {
    const {headers, data} = useCSVPreview({
      downloadLink: '/download/file.csv',
    });

    return (
      <>
        <span>headers: {JSON.stringify(headers)}</span>
        <span>data: {JSON.stringify(data)}</span>
      </>
    );
  });

  beforeAll(() => {
    server.listen();
  });

  afterAll(() => {
    server.close();
  });

  it('should get headers and data for a single line csv', async () => {
    server.use(
      rest.get(`/download/file.csv`, (_req, res, ctx) => {
        return res(ctx.text('a,b,c,d\n'));
      }),
    );
    render(<Component />);

    expect(
      await screen.findByText('headers: ["1","2","3","4"]'),
    ).toBeInTheDocument();
    expect(
      await screen.findByText('data: [{"1":"a","2":"b","3":"c","4":"d"}]'),
    ).toBeInTheDocument();
  });

  it('should get headers and data for a single line csv without a trailing newline', async () => {
    server.use(
      rest.get(`/download/file.csv`, (_req, res, ctx) => {
        return res(ctx.text('a,b,c,d'));
      }),
    );
    render(<Component />);

    expect(
      await screen.findByText('headers: ["1","2","3","4"]'),
    ).toBeInTheDocument();
    expect(
      await screen.findByText('data: [{"1":"a","2":"b","3":"c","4":"d"}]'),
    ).toBeInTheDocument();
  });

  it('should get headers and data for a multi line csv', async () => {
    server.use(
      rest.get(`/download/file.csv`, (_req, res, ctx) => {
        return res(ctx.text('name,age\njohn,25\njane,26\n'));
      }),
    );
    render(<Component />);

    expect(
      await screen.findByText('headers: ["name","age"]'),
    ).toBeInTheDocument();
    expect(
      await screen.findByText(
        'data: [{"name":"john","age":"25"},{"name":"jane","age":"26"}]',
      ),
    ).toBeInTheDocument();
  });

  it('should get headers and data for a multi line csv without a trailing newline', async () => {
    server.use(
      rest.get(`/download/file.csv`, (_req, res, ctx) => {
        return res(ctx.text('name,age\njohn,25\njane,26'));
      }),
    );
    render(<Component />);

    expect(
      await screen.findByText('headers: ["name","age"]'),
    ).toBeInTheDocument();
    expect(
      await screen.findByText(
        'data: [{"name":"john","age":"25"},{"name":"jane","age":"26"}]',
      ),
    ).toBeInTheDocument();
  });
});
