import {render} from '@testing-library/react';
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
      '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/txt_spec.txt',
    );
    const {findByText} = render(<FileBrowser />);
    expect(await findByText('name: visualizations')).toBeInTheDocument();
    expect(
      await findByText('image: elephantjones/market_sentiment:dev0.25'),
    ).toBeInTheDocument();
  });

  it('should support .jsonl file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/jsonl_people.jsonl',
    );
    const {findByText} = render(<FileBrowser />);
    expect(
      await findByText(
        '{"id":1,"father":"Mark","mother":"Charlotte","children":["Tom"]}',
      ),
    ).toBeInTheDocument();
    expect(
      await findByText(
        '{"id":3,"father":"Bob","mother":"Monika","children":["Jerry","Karol"]}',
      ),
    ).toBeInTheDocument();
  });

  it('should support .textpb file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/carriers_list.textpb',
    );
    const {findByText} = render(<FileBrowser />);
    expect(
      await findByText('carrier_name: "T-Mobile - US"'),
    ).toBeInTheDocument();
    expect(await findByText('mccmnc_tuple: "310026"')).toBeInTheDocument();
  });
});
