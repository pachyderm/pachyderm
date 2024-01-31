import React from 'react';
import {render, waitFor} from '@testing-library/react';

import * as requestAPI from '../../../../../handler';
import {mockedRequestAPI} from 'utils/testUtils';
import userEvent from '@testing-library/user-event';
import {AuthConfig} from 'plugins/mount/types';
import Config from '../Config';
jest.mock('../../../../../handler');

describe('config screen', () => {
  const mockRequestAPI = requestAPI as jest.Mocked<typeof requestAPI>;
  let updateConfig = jest.fn();
  const authConfig: AuthConfig = {
    cluster_status: 'INVALID',
  };

  beforeEach(() => {
    updateConfig = jest.fn();
    mockRequestAPI.requestAPI.mockImplementation(mockedRequestAPI({}));
  });

  describe('INVALID config', () => {
    it('should ask the user to provide a pachd address', () => {
      const {getByTestId, queryByTestId} = render(
        <Config
          updateConfig={updateConfig}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      getByTestId('Config__pachdAddressInput');
      expect(queryByTestId('Config__back')).not.toBeInTheDocument();
      expect(
        queryByTestId('Config__pachdAddressCancel'),
      ).not.toBeInTheDocument();
      expect(queryByTestId('Config__login')).not.toBeInTheDocument();
      expect(queryByTestId('Config__logout')).not.toBeInTheDocument();
    });
  });

  describe('AUTH_ENABLED config', () => {
    it('should show authenticated view', async () => {
      const authConfig: AuthConfig = {
        cluster_status: 'VALID_LOGGED_IN',
        pachd_address: 'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      };

      const {getByTestId, queryByTestId} = render(
        <Config
          updateConfig={updateConfig}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      expect(getByTestId('Config__pachdAddress')).toHaveTextContent(
        'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      );
      getByTestId('Config__pachdAddressUpdate');
      queryByTestId('Config__logout');
      expect(queryByTestId('Config__login')).not.toBeInTheDocument();
    });

    it('should show unauthenticated view', () => {
      const authConfig: AuthConfig = {
        cluster_status: 'VALID_LOGGED_OUT',
        pachd_address: 'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      };

      const {getByTestId, queryByTestId} = render(
        <Config
          updateConfig={updateConfig}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      expect(queryByTestId('Config__back')).not.toBeInTheDocument();
      expect(getByTestId('Config__pachdAddress')).toHaveTextContent(
        'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      );
      getByTestId('Config__pachdAddressUpdate');
      queryByTestId('Config__login');
      expect(queryByTestId('Config__logout')).not.toBeInTheDocument();
    });

    it('should allow user to logout', async () => {
      const authConfig: AuthConfig = {
        cluster_status: 'VALID_LOGGED_IN',
        pachd_address: 'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      };

      const {findByTestId} = render(
        <Config
          updateConfig={updateConfig}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      (await findByTestId('Config__logout')).click();

      await waitFor(() => {
        expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
          'auth/_logout',
          'PUT',
        );
      });
    });

    it('should allow user to login', async () => {
      const authConfig: AuthConfig = {
        cluster_status: 'VALID_LOGGED_OUT',
        pachd_address: 'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      };

      window.open = jest.fn();
      const loginUrl =
        'https://hub-c0-jwn7iwcca9.clusters.pachyderm.io/dex/auth?client_id=pachd';
      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({loginUrl: loginUrl}),
      );

      const {findByTestId} = render(
        <Config
          updateConfig={updateConfig}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      (await findByTestId('Config__login')).click();

      await waitFor(() => {
        expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
          'auth/_login',
          'PUT',
        );
      });

      expect(window.open).toHaveBeenCalledWith(loginUrl, '', expect.anything());
    });
  });

  describe('AUTH_DISABLED config', () => {
    it('should display default view', () => {
      const authConfig: AuthConfig = {
        cluster_status: 'VALID_NO_AUTH',
        pachd_address: 'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      };

      const {getByTestId, queryByTestId} = render(
        <Config
          updateConfig={updateConfig}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      expect(getByTestId('Config__pachdAddress')).toHaveTextContent(
        'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      );
      getByTestId('Config__pachdAddressUpdate');
      expect(queryByTestId('Config__login')).not.toBeInTheDocument();
      expect(queryByTestId('Config__logout')).not.toBeInTheDocument();
    });
  });

  describe('pachd address field', () => {
    it('should validate entered address', async () => {
      const authConfig: AuthConfig = {
        cluster_status: 'INVALID',
      };

      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          cluster_status: 'INVALID',
        }),
      );

      const {getByTestId, findByText} = render(
        <Config
          updateConfig={updateConfig}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      const input = getByTestId('Config__pachdAddressInput');
      const submit = getByTestId('Config__pachdAddressSubmit');

      userEvent.type(input, 'grpc://test.com:31400');
      submit.click();
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(1);

      userEvent.clear(input);
      userEvent.type(input, 'grpcs://test.com:31400');
      submit.click();
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(2);

      userEvent.clear(input);
      userEvent.type(input, 'http://test.com:31400');
      submit.click();
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(3);

      userEvent.clear(input);
      userEvent.type(input, 'https://test.com:31400');
      submit.click();
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(4);

      userEvent.clear(input);
      userEvent.type(input, 'unix://test.com:31400');
      submit.click();
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(5);

      userEvent.clear(input);
      userEvent.type(input, 'www.test.com');
      submit.click();
      await findByText(
        'Cluster address should start with grpc://, grpcs://, http://, https:// or unix://',
      );
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(5);
    });

    it('should allow user to update config', async () => {
      const authConfig: AuthConfig = {
        cluster_status: 'VALID_LOGGED_IN',
        pachd_address: 'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      };

      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          cluster_status: 'VALID_LOGGED_IN',
          pachd_address:
            'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
        }),
      );

      const {getByTestId} = render(
        <Config
          updateConfig={updateConfig}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      getByTestId('Config__pachdAddressUpdate').click();
      expect(getByTestId('Config__pachdAddressSubmit')).toBeDisabled();

      const input = getByTestId('Config__pachdAddressInput');
      userEvent.type(
        input,
        'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
      );
      getByTestId('Config__pachdAddressSubmit').click();

      await waitFor(() => {
        expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
          'config',
          'PUT',
          {
            pachd_address:
              'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
          },
        );
        expect(updateConfig).toHaveBeenCalledTimes(1);
      });
    });

    it('should allow user to set advanced config options', async () => {
      const authConfig: AuthConfig = {
        cluster_status: 'VALID_LOGGED_IN',
        pachd_address: 'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      };

      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          cluster_status: 'VALID_LOGGED_IN',
          pachd_address:
            'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
        }),
      );

      const {getByTestId} = render(
        <Config
          updateConfig={updateConfig}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      getByTestId('Config__pachdAddressUpdate').click();
      expect(getByTestId('Config__pachdAddressSubmit')).toBeDisabled();

      getByTestId('Config__advancedSettingsToggle').click();
      const textArea = getByTestId('Config__serverCaInput');
      userEvent.type(textArea, '12345=');

      expect(getByTestId('Config__pachdAddressSubmit')).toBeDisabled();

      const input = getByTestId('Config__pachdAddressInput');
      userEvent.type(
        input,
        'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
      );
      getByTestId('Config__pachdAddressSubmit').click();

      await waitFor(() => {
        expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
          'config',
          'PUT',
          {
            pachd_address:
              'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
            server_cas: '12345=',
          },
        );
        expect(updateConfig).toHaveBeenCalledTimes(1);
      });
    });
  });
});
