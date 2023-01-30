import React from 'react';
import {render, waitFor} from '@testing-library/react';
import {ServerConnection} from '@jupyterlab/services';
import userEvent from '@testing-library/user-event';
import YAML from 'yaml';

import * as requestAPI from '../../../../../handler';
import {mockedRequestAPI} from 'utils/testUtils';
import Datum from '../Datum';
jest.mock('../../../../../handler');

describe('datum screen', () => {
  let setShowDatum = jest.fn();
  let setKeepMounted = jest.fn();
  const mockRequestAPI = requestAPI as jest.Mocked<typeof requestAPI>;

  beforeEach(() => {
    setShowDatum = jest.fn();
    setKeepMounted = jest.fn();
    mockRequestAPI.requestAPI.mockImplementation(mockedRequestAPI({}));

    // IntersectionObserver isn't available in test environment
    const mockIntersectionObserver = jest.fn();
    mockIntersectionObserver.mockReturnValue({
      observe: () => null,
      unobserve: () => null,
      disconnect: () => null,
    });
    window.IntersectionObserver = mockIntersectionObserver;
  });

  describe('mounting datums', () => {
    it('successful mount datums call shows cycler', async () => {
      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          id: 'asdfaew34ri92jafiolwe',
          idx: 0,
          num_datums: 6,
        }),
      );

      const {getByTestId, queryByTestId, findByTestId} = render(
        <Datum
          showDatum={true}
          setShowDatum={setShowDatum}
          keepMounted={false}
          setKeepMounted={setKeepMounted}
          open={jest.fn()}
          pollRefresh={jest.fn()}
          repoViewInputSpec={{}}
        />,
      );

      expect(queryByTestId('Datum__cyclerLeft')).not.toBeInTheDocument();
      expect(queryByTestId('Datum__cyclerRight')).not.toBeInTheDocument();

      const input = await findByTestId('Datum__inputSpecInput');
      const submit = await findByTestId('Datum__mountDatums');

      userEvent.type(input, '{"pfs": "a"}'.replace(/[{[]/g, '$&$&'));
      expect(input).toHaveValue('{"pfs": "a"}');
      submit.click();

      await waitFor(() => {
        expect(mockRequestAPI.requestAPI).toHaveBeenNthCalledWith(
          2,
          '_mount_datums',
          'PUT',
          {input: {pfs: 'a'}},
        );
      });

      getByTestId('Datum__cyclerLeft');
      getByTestId('Datum__cyclerRight');
      expect(getByTestId('Datum__cycler')).toHaveTextContent('(1/6)');
    });
  });

  describe('cycle through datums', () => {
    it('hitting datum cycler makes show datum call', async () => {
      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          id: 'asdfaew34ri92jafiolwe',
          idx: 0,
          num_datums: 6,
        }),
      );

      const {getByTestId, findByTestId} = render(
        <Datum
          showDatum={true}
          setShowDatum={setShowDatum}
          keepMounted={false}
          setKeepMounted={setKeepMounted}
          open={jest.fn()}
          pollRefresh={jest.fn()}
          repoViewInputSpec={{}}
        />,
      );

      const input = await findByTestId('Datum__inputSpecInput');
      const submit = await findByTestId('Datum__mountDatums');

      userEvent.type(input, '{"pfs": "a"}'.replace(/[{[]/g, '$&$&'));
      expect(input).toHaveValue('{"pfs": "a"}');
      await submit.click();

      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          id: 'ilwe9nme9902ja039jf20snv',
          idx: 1,
          num_datums: 6,
        }),
      );

      getByTestId('Datum__cyclerLeft');
      (await findByTestId('Datum__cyclerRight')).click();

      await waitFor(() => {
        expect(mockRequestAPI.requestAPI).toHaveBeenNthCalledWith(
          4,
          '_show_datum?idx=1',
          'PUT',
        );
      });

      getByTestId('Datum__cyclerLeft');
      getByTestId('Datum__cyclerRight');
      expect(getByTestId('Datum__cycler')).toHaveTextContent('(2/6)');
    });
  });

  describe('errors with input spec', () => {
    it('error if bad syntax in input spec', async () => {
      const {getByTestId, findByTestId} = render(
        <Datum
          showDatum={true}
          setShowDatum={setShowDatum}
          keepMounted={false}
          setKeepMounted={setKeepMounted}
          open={jest.fn()}
          pollRefresh={jest.fn()}
          repoViewInputSpec={{}}
        />,
      );

      expect(getByTestId('Datum__errorMessage')).toHaveTextContent('');

      const input = await findByTestId('Datum__inputSpecInput');
      const submit = await findByTestId('Datum__mountDatums');

      userEvent.type(input, '{"pfs": "a"'.replace(/[{[]/g, '$&$&'));
      expect(input).toHaveValue('{"pfs": "a"');
      submit.click();

      expect(getByTestId('Datum__errorMessage')).toHaveTextContent(
        'Poorly formatted input spec',
      );
    });

    it('error if invalid references in input spec', async () => {
      mockRequestAPI.requestAPI.mockImplementation(() => {
        throw new ServerConnection.ResponseError(new Response());
      });

      const {getByTestId, findByTestId} = render(
        <Datum
          showDatum={true}
          setShowDatum={setShowDatum}
          keepMounted={false}
          setKeepMounted={setKeepMounted}
          open={jest.fn()}
          pollRefresh={jest.fn()}
          repoViewInputSpec={{}}
        />,
      );

      expect(getByTestId('Datum__errorMessage')).toHaveTextContent('');

      const input = await findByTestId('Datum__inputSpecInput');
      const submit = await findByTestId('Datum__mountDatums');

      userEvent.type(input, '{"pfs": "fake_repo"}'.replace(/[{[]/g, '$&$&'));
      expect(input).toHaveValue('{"pfs": "fake_repo"}');
      submit.click();

      expect(getByTestId('Datum__errorMessage')).toHaveTextContent(
        'Bad data in input spec',
      );
    });
  });

  describe('test valid input spec formats', () => {
    it('valid json input spec', async () => {
      const {getByTestId, findByTestId} = render(
        <Datum
          showDatum={true}
          setShowDatum={setShowDatum}
          keepMounted={false}
          setKeepMounted={setKeepMounted}
          open={jest.fn()}
          pollRefresh={jest.fn()}
          repoViewInputSpec={{}}
        />,
      );

      expect(getByTestId('Datum__errorMessage')).toHaveTextContent('');

      const input = await findByTestId('Datum__inputSpecInput');
      const submit = await findByTestId('Datum__mountDatums');

      userEvent.type(input, '{"pfs": "repo"}'.replace(/[{[]/g, '$&$&'));
      expect(input).toHaveValue('{"pfs": "repo"}');
      submit.click();

      expect(getByTestId('Datum__errorMessage')).toHaveTextContent('');
    });

    it('valid yaml input spec', async () => {
      const {getByTestId, findByTestId} = render(
        <Datum
          showDatum={true}
          setShowDatum={setShowDatum}
          keepMounted={false}
          setKeepMounted={setKeepMounted}
          open={jest.fn()}
          pollRefresh={jest.fn()}
          repoViewInputSpec={{}}
        />,
      );

      expect(getByTestId('Datum__errorMessage')).toHaveTextContent('');

      const input = await findByTestId('Datum__inputSpecInput');
      const submit = await findByTestId('Datum__mountDatums');

      userEvent.type(input, YAML.stringify({pfs: 'repo'}));
      expect(input).toHaveValue(YAML.stringify({pfs: 'repo'}));
      submit.click();

      expect(getByTestId('Datum__errorMessage')).toHaveTextContent('');
    });
  });
});
