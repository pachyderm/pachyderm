import React from 'react';
import {KubernetesElephantSVG} from '@pachyderm/components';
import {AuthConfig} from 'plugins/mount/types';
import {useConfig} from './hooks/useConfig';
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
  const {
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
  } = useConfig(
    showConfig,
    setShowConfig,
    updateConfig,
    authConfig,
    refresh,
    reposStatus,
  );

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
                placeholder="grpcs://example.pachyderm.com:30650"
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
