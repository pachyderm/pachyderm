import {
  render,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import FileBrowserComponent from '@dash-frontend/views/FileBrowser';

import CodePreview from '../CodePreview';

describe('Code Preview', () => {
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });

  it('should support markdown files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/markdown_basic.md',
    );
    render(<FileBrowser />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(await screen.findByText('View Source')).toBeInTheDocument();
    await userEvent.click(screen.getByTestId('Switch__button'));

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(await screen.findByText('###')).toBeInTheDocument();
  });

  it('should support .json files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/json_mixed.json',
    );
    render(<FileBrowser />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(
      await screen.findByText((text) => text.includes('231221B')),
    ).toBeInTheDocument();
  });

  it('should support .yml files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/yml_spec.yml',
    );
    render(<FileBrowser />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(
      await screen.findByText((text) =>
        text.includes('elephantjones/market_sentiment'),
      ),
    ).toBeInTheDocument();
  });

  it('should support .txt files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/txt_spec.txt',
    );
    render(<FileBrowser />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(
      await screen.findByText((text) => text.includes('name: visualizations')),
    ).toBeInTheDocument();
  });

  it('should support .jsonl files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/jsonl_people.jsonl',
    );
    render(<FileBrowser />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(
      await screen.findByText((text) =>
        text.includes(
          '{"id":1,"father":"Mark","mother":"Charlotte","children":["Tom"]}',
        ),
      ),
    ).toBeInTheDocument();
  });

  it('should support .textpb files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/carriers_list.textpb',
    );
    render(<FileBrowser />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(
      await screen.findByText((text) =>
        text.includes('carrier_name: "T-Mobile - US"'),
      ),
    ).toBeInTheDocument();
  });

  it('should support rendering from source', async () => {
    render(<CodePreview source="Hello World" />);

    expect(
      await screen.findByText((text) => text.includes('Hello World')),
    ).toBeInTheDocument();
  });

  it('should render different style modes', async () => {
    render(
      <CodePreview source="Modes" hideGutter hideLineNumbers fullHeight />,
    );

    const wrapper = screen.getByTestId('CodePreview__wrapper');

    expect(wrapper).toHaveClass('hideGutter');
    expect(wrapper).toHaveClass('hideLineNumbers');
    expect(wrapper).toHaveClass('fullHeight');
  });
});
