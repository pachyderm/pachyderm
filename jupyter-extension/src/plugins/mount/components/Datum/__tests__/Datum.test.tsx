import React from 'react';
import {render, waitFor} from '@testing-library/react';

import * as requestAPI from '../../../../../handler';
import {mockedRequestAPI} from 'utils/testUtils';
import userEvent from '@testing-library/user-event';
import Datum from '../Datum';
jest.mock('../../../../../handler');

describe('datum screen', () => {
  const mockRequestAPI = requestAPI as jest.Mocked<typeof requestAPI>;
  let setShowDatum = jest.fn();
  let setKeepMounted = jest.fn();

  beforeEach(() => {
    setShowDatum = jest.fn();
    setKeepMounted = jest.fn();
    mockRequestAPI.requestAPI.mockImplementation(mockedRequestAPI({}));
  });

  it('should allow users to mount datums', async () => {
    mockRequestAPI.requestAPI.mockImplementation(
      mockedRequestAPI({
        id: 'asdfaew34ri92jafiolwe',
        idx: 0,
        num_datums: 6,
      }),
    );

    const {getByTestId, queryByTestId} = render(
      <Datum
        showDatum={true}
        setShowDatum={setShowDatum}
        keepMounted={false}
        setKeepMounted={setKeepMounted}
        refresh={jest.fn()}
        pollRefresh={jest.fn()}
      />,
    );

    expect(queryByTestId('Datum__cyclerLeft')).not.toBeInTheDocument();
    expect(queryByTestId('Datum__cyclerRight')).not.toBeInTheDocument();

    const input = getByTestId('Datum__inputSpecInput');
    const submit = getByTestId('Datum__mountDatums');
    userEvent.type(input, '{"pfs": "a"}'.replace(/[{[]/g, '$&$&'));
    submit.click();

    await waitFor(() => {
      expect(mockRequestAPI.requestAPI).nthCalledWith(
        2,
        '_mount_datums',
        'PUT',
        {input: {pfs: 'a'}},
      );
    });

    getByTestId('Datum__cyclerLeft');
    getByTestId('Datum__cyclerRight');
  });
});

// show error for invalid JSON
// show error for bad data in spec
// mount datums button makes mount datums call; show cycler after mounting datums
// cycler button makes show datum call
