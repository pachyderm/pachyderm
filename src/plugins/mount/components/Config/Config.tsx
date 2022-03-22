import {requestAPI} from '../../../../handler';
import React, {useEffect, useState} from 'react';
import {KubernetesElephantSVG, usePreviousValue} from '@pachyderm/components';
import {AuthConfig} from 'plugins/mount/types';
type ConfigProps = {
  showConfig: boolean;
  setShowConfig: (shouldShow: boolean) => void;
  reposStatus?: number;
  updateConfig: (shouldShow: AuthConfig) => void;
  authConfig: AuthConfig;
  refresh: () => Promise<void>;
};

const Config: React.FC<ConfigProps> = ({
  showConfig,
  setShowConfig,
  reposStatus,
  updateConfig,
  authConfig,
  refresh,
}) => {
  const [loading, setLoading] = useState(false);
  const [addressField, setAddressField] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [shouldShowLogin, setShouldShowLogin] = useState(false);
  const [shouldShowAddressInput, setShouldShowAddressInput] = useState(false);

  useEffect(() => {
    if (showConfig) {
      setShouldShowLogin(authConfig.cluster_status === 'AUTH_ENABLED');
      setShouldShowAddressInput(authConfig.cluster_status === 'INVALID');
    }
    setErrorMessage('');
    setAddressField('');
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
        const response = await requestAPI<AuthConfig>('config', 'PUT', {
          pachd_address: tmpAddress,
        });

        if (response.cluster_status === 'INVALID') {
          setErrorMessage('Invalid address.');
        } else {
          updateConfig(response);
          setShouldShowLogin(response.cluster_status === 'AUTH_ENABLED');
        }
      } else {
        setErrorMessage(
          'address should start with grpc://, grpcs://, http://, https:// or unix://',
        );
      }
    } catch (e) {
      setErrorMessage('error setting pachd address.');
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
  return (
    <>
      <div className="pachyderm-mount-config-form-base">
        {reposStatus === 200 && (
          <button
            data-testid="Config__back"
            className="pachyderm-button"
            onClick={() => setShowConfig(false)}
          >
            Back
          </button>
        )}
        <div className="pachyderm-mount-config-heading">
          Pachyderm
          <span className="pachyderm-mount-config-subheading">
            Mount Extension
          </span>
        </div>
        <div className="pachyderm-mount-config-address-wrapper">
          {authConfig.pachd_address && !shouldShowAddressInput ? (
            <>
              <label
                htmlFor="pachd"
                className="pachyderm-mount-config-address-label"
              >
                Cluster Address
              </label>
              <span
                data-testid="Config__pachdAddress"
                className="pachyderm-mount-config-address"
              >
                {authConfig.pachd_address}
              </span>
              <button
                data-testid="Config__pachdAddressUpdate"
                className="pachyderm-button"
                onClick={() => setShouldShowAddressInput(true)}
              >
                Change Address
              </button>
            </>
          ) : (
            <>
              <label
                htmlFor="pachd"
                className="pachyderm-mount-config-address-label"
              >
                Cluster Address
              </label>
              <input
                data-testid="Config__pachdAddressInput"
                name="pachd"
                value={addressField}
                onInput={(e: any) => {
                  if (errorMessage) {
                    setErrorMessage('');
                  }
                  setAddressField(e.target.value);
                }}
                placeholder="grpcs://example.pachyderm.com:30600"
              ></input>
              <span className="pachyderm-mount-config-address-error">
                {errorMessage}
              </span>
              <button
                data-testid="Config__pachdAddressSubmit"
                className="pachyderm-button pachyderm-mount-config-set-address"
                disabled={loading || !addressField}
                onClick={updatePachdAddress}
              >
                Set Address
              </button>

              {authConfig.pachd_address && (
                <button
                  data-testid="Config__pachdAddressCancel"
                  className="pachyderm-button"
                  disabled={loading}
                  onClick={() => {
                    setErrorMessage('');
                    setAddressField('');
                    setShouldShowAddressInput(false);
                  }}
                >
                  Cancel
                </button>
              )}
            </>
          )}
        </div>
        {shouldShowLogin && !shouldShowAddressInput && (
          <div className="pachyderm-mount-login-container">
            {reposStatus === 200 ? (
              <button
                data-testid="Config__logout"
                className="pachyderm-button"
                disabled={loading}
                onClick={callLogout}
              >
                Logout
              </button>
            ) : (
              <button
                data-testid="Config__login"
                className="pachyderm-button"
                disabled={loading}
                onClick={callLogin}
              >
                Login
              </button>
            )}
          </div>
        )}
      </div>
      <div className="pachyderm-mount-config-graphic-base">
        <div className="pachyderm-mount-config-graphic-container">
          <KubernetesElephantSVG
            width="230px"
            height="230px"
            viewBox="170 0 400 400"
          />
        </div>
      </div>
    </>
  );
};

export default Config;
