import React from 'react';
import {act, render, screen, waitFor} from '@testing-library/react';

import * as requestAPI from '../../../../../handler';
import {mockedRequestAPI} from 'utils/testUtils';
import userEvent from '@testing-library/user-event';
import {AuthConfig, HealthCheck} from 'plugins/mount/types';
import Config from '../Config';
jest.mock('../../../../../handler');

describe('config screen', () => {
  const mockRequestAPI = requestAPI as jest.Mocked<typeof requestAPI>;
  let updateConfig = jest.fn();
  const healthCheck: HealthCheck = {
    status: 'HEALTHY_INVALID_CLUSTER',
  };
  const authConfig: AuthConfig = {
    pachd_address: 'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
  };

  beforeEach(() => {
    updateConfig = jest.fn();
    mockRequestAPI.requestAPI.mockImplementation(mockedRequestAPI({}));
  });

  describe('INVALID config', () => {
    it('should ask the user to provide a pachd address', () => {
      const authConfig: AuthConfig = {};
      const {getByTestId, queryByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
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
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_LOGGED_IN',
      };

      const {getByTestId, queryByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      getByTestId('Config__pachdAddress');

      expect(getByTestId('Config__pachdAddress')).toHaveTextContent(
        'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      );
      getByTestId('Config__pachdAddressUpdate');
      queryByTestId('Config__logout');
      expect(queryByTestId('Config__login')).not.toBeInTheDocument();
    });

    it('should show unauthenticated view', () => {
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_LOGGED_OUT',
      };

      const {getByTestId, queryByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
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
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_LOGGED_IN',
      };

      const {findByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      await act(async () => {
        (await findByTestId('Config__logout')).click();
      });

      await waitFor(() => {
        expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
          'auth/_logout',
          'PUT',
        );
      });
    });

    it('logged in should display option to change address', async () => {
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_LOGGED_IN',
      };

      const {findByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      expect(
        await screen.findByTestId('Config__pachdAddressUpdate'),
      ).toBeInTheDocument();
      await act(async () => {
        (await findByTestId('Config__pachdAddressUpdate')).click();
      });
      expect(
        await screen.findByTestId('Config__mountConfigSubheading'),
      ).toHaveTextContent('Update Configuration');
      await act(async () => {
        (await findByTestId('Config__pachdAddressCancel')).click();
      });
      expect(
        await screen.findByTestId('Config__pachdAddress'),
      ).toHaveTextContent(
        'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      );
    });

    it('should allow user to login', async () => {
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_LOGGED_OUT',
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
          healthCheck={healthCheck}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      await act(async () => {
        (await findByTestId('Config__login')).click();
      });

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
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_NO_AUTH',
      };

      const {getByTestId, queryByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
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

    it('should display option to change address', async () => {
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_NO_AUTH',
      };

      const {getByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      await act(async () => {
        await getByTestId('Config__pachdAddressUpdate').click();
      });
      expect(getByTestId('Config__mountConfigSubheading')).toHaveTextContent(
        'Update Configuration',
      );
      await act(async () => {
        await getByTestId('Config__pachdAddressCancel').click();
      });
      expect(getByTestId('Config__pachdAddress')).toHaveTextContent(
        'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      );
    });

    it('should show address after address is resubmitted', async () => {
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_NO_AUTH',
      };

      const {getByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          pachd_address:
            'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
        }),
      );

      expect(getByTestId('Config__pachdAddress')).toHaveTextContent(
        'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      );
      await act(async () => {
        await getByTestId('Config__pachdAddressUpdate').click();
      });
      expect(getByTestId('Config__mountConfigSubheading')).toHaveTextContent(
        'Update Configuration',
      );
      const input = getByTestId('Config__pachdAddressInput');
      userEvent.type(
        input,
        'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      );
      await act(async () => {
        await getByTestId('Config__pachdAddressSubmit').click();
      });

      await waitFor(() => {
        expect(mockRequestAPI.requestAPI).toHaveBeenCalledWith(
          'config',
          'PUT',
          {
            pachd_address:
              'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
          },
        );
        expect(updateConfig).toHaveBeenCalledTimes(1);
      });
      expect(getByTestId('Config__pachdAddress')).toHaveTextContent(
        'grpcs://hub-c0-jwn7iwcca9.clusters.pachyderm.io:31400',
      );
    });
  });

  describe('pachd address field', () => {
    it('should validate entered address', async () => {
      const authConfig: AuthConfig = {
        pachd_address: '',
      };

      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          status: 'HEALTHY_INVALID_CLUSTER',
        }),
      );

      const {getByTestId, findByText} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      const input = getByTestId('Config__pachdAddressInput');
      const submit = getByTestId('Config__pachdAddressSubmit');

      userEvent.type(input, 'grpc://test.com:31400');
      await act(async () => {
        submit.click();
      });
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(1);

      userEvent.clear(input);
      userEvent.type(input, 'grpcs://test.com:31400');
      await act(async () => {
        submit.click();
      });
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(2);

      userEvent.clear(input);
      userEvent.type(input, 'http://test.com:31400');
      await act(async () => {
        submit.click();
      });
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(3);

      userEvent.clear(input);
      userEvent.type(input, 'https://test.com:31400');
      await act(async () => {
        submit.click();
      });
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(4);

      userEvent.clear(input);
      userEvent.type(input, 'unix://test.com:31400');
      await act(async () => {
        submit.click();
      });
      await findByText('Invalid address.');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(5);

      userEvent.clear(input);
      userEvent.type(input, 'www.test.com');
      await act(async () => {
        submit.click();
      });
      await findByText(
        'Cluster address should start with grpc://, grpcs://, http://, https:// or unix://',
      );
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(5);
    });

    it('should allow user to update config', async () => {
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_LOGGED_IN',
      };

      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          pachd_address:
            'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
        }),
      );

      const {getByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      await act(async () => {
        getByTestId('Config__pachdAddressUpdate').click();
      });
      expect(getByTestId('Config__pachdAddressSubmit')).toBeDisabled();

      const input = getByTestId('Config__pachdAddressInput');
      userEvent.type(
        input,
        'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
      );
      await act(async () => {
        getByTestId('Config__pachdAddressSubmit').click();
      });

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
      const healthCheck: HealthCheck = {
        status: 'HEALTHY_LOGGED_IN',
      };

      mockRequestAPI.requestAPI.mockImplementation(
        mockedRequestAPI({
          pachd_address:
            'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
        }),
      );

      const {getByTestId} = render(
        <Config
          updateConfig={updateConfig}
          healthCheck={healthCheck}
          authConfig={authConfig}
          refresh={jest.fn()}
        />,
      );

      await act(async () => {
        getByTestId('Config__pachdAddressUpdate').click();
      });
      expect(getByTestId('Config__pachdAddressSubmit')).toBeDisabled();

      await act(async () => {
        getByTestId('Config__advancedSettingsToggle').click();
      });
      const textArea = getByTestId('Config__serverCaInput');
      userEvent.type(textArea, '12345=');

      expect(getByTestId('Config__pachdAddressSubmit')).toBeDisabled();

      const input = getByTestId('Config__pachdAddressInput');
      userEvent.type(
        input,
        'grpcs://hub-123-123123123.clusters.pachyderm.io:31400',
      );
      await act(async () => {
        getByTestId('Config__pachdAddressSubmit').click();
      });

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
