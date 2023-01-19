import {render, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import FileBrowserComponent from '@dash-frontend/views/FileBrowser';

describe('Text Preview', () => {
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });

  it('should support .txt file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/txt_spec.txt',
    );
    render(<FileBrowser />);
    expect(await screen.findByText('name: visualizations')).toBeInTheDocument();
    expect(
      await screen.findByText('image: elephantjones/market_sentiment:dev0.25'),
    ).toBeInTheDocument();
  });

  it('should support .jsonl file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/jsonl_people.jsonl',
    );
    render(<FileBrowser />);
    expect(
      await screen.findByText(
        '{"id":1,"father":"Mark","mother":"Charlotte","children":["Tom"]}',
      ),
    ).toBeInTheDocument();
    expect(
      await screen.findByText(
        '{"id":3,"father":"Bob","mother":"Monika","children":["Jerry","Karol"]}',
      ),
    ).toBeInTheDocument();
  });

  it('should support .textpb file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/carriers_list.textpb',
    );
    render(<FileBrowser />);
    expect(
      await screen.findByText('carrier_name: "T-Mobile - US"'),
    ).toBeInTheDocument();
    expect(
      await screen.findByText('mccmnc_tuple: "310026"'),
    ).toBeInTheDocument();
  });
});
