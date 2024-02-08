import {requestAPI} from '../../../../../handler';
import {useEffect, useState} from 'react';
import {AuthConfig, clusterStatus} from 'plugins/mount/types';

export type useConfigResponse = {
  addressField: string;
  setAddressField: (address: string) => void;
  errorMessage: string;
  setErrorMessage: (message: string) => void;
  shouldShowAddressInput: boolean;
  setShouldShowAddressInput: (show: boolean) => void;
  updatePachdAddress: () => Promise<void>;
  callLogin: () => Promise<void>;
  callLogout: () => Promise<void>;
  clusterStatus: clusterStatus;
  loading: boolean;
  showAdvancedOptions: boolean;
  setShowAdvancedOptions: (show: boolean) => void;
  serverCa: string;
  setServerCa: (serverCa: string) => void;
};

export const useConfig = (
  updateConfig: (shouldShow: AuthConfig) => void,
  authConfig: AuthConfig,
  refresh: () => Promise<void>,
): useConfigResponse => {
  const [loading, setLoading] = useState(false);
  const [addressField, setAddressField] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [clusterStatus, setClusterStatus] = useState('NONE' as clusterStatus);
  const [shouldShowAddressInput, setShouldShowAddressInput] = useState(false);
  const [showAdvancedOptions, setShowAdvancedOptions] = useState(false);
  const [serverCa, setServerCa] = useState('');

  useEffect(() => {
    setClusterStatus(authConfig.cluster_status);
    setShouldShowAddressInput(
      ['NONE', 'UNKNOWN', 'INVALID'].includes(authConfig.cluster_status),
    );
    setErrorMessage('');
    setAddressField('');
    setServerCa('');
    setShowAdvancedOptions(false);
  }, [authConfig]);

  const updatePachdAddress = async () => {
    setLoading(true);
    try {
      const tmpAddress = addressField.trim();
      const validAddressPattern = /^((grpc|grpcs|http|https|unix):\/\/)/;

      if (validAddressPattern.test(tmpAddress)) {
        const response = await requestAPI<AuthConfig>(
          'config',
          'PUT',
          serverCa
            ? {
                pachd_address: tmpAddress,
                server_cas: serverCa,
              }
            : {pachd_address: tmpAddress},
        );

        if (response.cluster_status === 'INVALID') {
          setErrorMessage('Invalid address.');
        } else if (response.cluster_status === 'UNKNOWN') {
          setErrorMessage(
            'An unexpected error occurred when attempting to set the address.',
          );
        } else {
          updateConfig(response);
        }
      } else {
        setErrorMessage(
          'Cluster address should start with grpc://, grpcs://, http://, https:// or unix://',
        );
      }
    } catch (e) {
      setErrorMessage('Unable to connect to cluster.');
      console.log(e);
    }
    setLoading(false);
  };

  const callLogin = async () => {
    setLoading(true);
    try {
      const res = await requestAPI<any>('auth/_login', 'PUT');
      if (res.loginUrl) {
        const x = window.screenX + (window.outerWidth - 500) / 2;
        const y = window.screenY + (window.outerHeight - 500) / 2.5;
        const features = `width=${500},height=${500},left=${x},top=${y}`;
        window.open(res.loginUrl, '', features);
      }
    } catch (e) {
      console.log(e);
    }

    // There is no current way to get infromation from the auth_url window.
    // Adding a timeout to prevent users from spamming the button.
    setTimeout(() => {
      setLoading(false);
    }, 2000);
    await refresh();
  };

  const callLogout = async () => {
    setLoading(true);
    try {
      await requestAPI<any>('auth/_logout', 'PUT');
      await refresh();
    } catch (e) {
      console.log(e);
    }
    setLoading(false);
  };

  return {
    addressField,
    setAddressField,
    errorMessage,
    setErrorMessage,
    shouldShowAddressInput,
    setShouldShowAddressInput,
    updatePachdAddress,
    callLogin,
    callLogout,
    clusterStatus,
    loading,
    showAdvancedOptions,
    setShowAdvancedOptions,
    serverCa,
    setServerCa,
  };
};
