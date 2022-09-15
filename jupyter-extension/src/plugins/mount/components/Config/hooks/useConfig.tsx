import {requestAPI} from '../../../../../handler';
import {useEffect, useState} from 'react';
import usePreviousValue from '../../../../../utils/hooks/usePreviousValue';
import {AuthConfig} from 'plugins/mount/types';

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
  shouldShowLogin: boolean;
  loading: boolean;
  showAdvancedOptions: boolean;
  setShowAdvancedOptions: (show: boolean) => void;
  serverCa: string;
  setServerCa: (serverCa: string) => void;
};

export const useConfig = (
  showConfig: boolean,
  setShowConfig: (shouldShow: boolean) => void,
  updateConfig: (shouldShow: AuthConfig) => void,
  authConfig: AuthConfig,
  refresh: () => Promise<void>,
  reposStatus?: number,
): useConfigResponse => {
  const [loading, setLoading] = useState(false);
  const [addressField, setAddressField] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [shouldShowLogin, setShouldShowLogin] = useState(false);
  const [shouldShowAddressInput, setShouldShowAddressInput] = useState(false);
  const [showAdvancedOptions, setShowAdvancedOptions] = useState(false);
  const [serverCa, setServerCa] = useState('');

  useEffect(() => {
    if (showConfig) {
      setShouldShowLogin(authConfig.cluster_status === 'AUTH_ENABLED');
      setShouldShowAddressInput(authConfig.cluster_status === 'INVALID');
    }
    setErrorMessage('');
    setAddressField('');
    setServerCa('');
    setShowAdvancedOptions(false);
  }, [showConfig, authConfig]);

  const previousStatus = usePreviousValue(reposStatus);

  useEffect(() => {
    if (
      showConfig &&
      previousStatus &&
      previousStatus !== 200 &&
      reposStatus === 200
    ) {
      setShowConfig(false);
    }
  }, [showConfig, reposStatus]);

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
        } else {
          updateConfig(response);
          setShouldShowLogin(response.cluster_status === 'AUTH_ENABLED');
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
      if (res.auth_url) {
        const x = window.screenX + (window.outerWidth - 500) / 2;
        const y = window.screenY + (window.outerHeight - 500) / 2.5;
        const features = `width=${500},height=${500},left=${x},top=${y}`;
        window.open(res.auth_url, '', features);
      }
    } catch (e) {
      console.log(e);
    }

    // There is no current way to get infromation from the auth_url window.
    // Adding a timeout to prevent users from spamming the button.
    setTimeout(() => {
      setLoading(false);
    }, 2000);
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
    shouldShowLogin,
    loading,
    showAdvancedOptions,
    setShowAdvancedOptions,
    serverCa,
    setServerCa,
  };
};
