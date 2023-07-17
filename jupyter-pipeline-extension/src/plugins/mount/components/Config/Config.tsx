import React from 'react';
import {AuthConfig} from 'plugins/mount/types';
import {useConfig} from './hooks/useConfig';
import {infoIcon} from '../../../../utils/icons';

import {closeIcon} from '@jupyterlab/ui-components';
import LoadingDots from '../../../../utils/components/LoadingDots/LoadingDots';
import {KubernetesElephant} from '../../../../utils/components/Svgs';

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
    showAdvancedOptions,
    setShowAdvancedOptions,
    serverCa,
    setServerCa,
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
          <div className="pachyderm-mount-config-back">
            <button
              data-testid="Config__back"
              className="pachyderm-button-link"
              onClick={() => setShowConfig(false)}
            >
              Back{' '}
              <closeIcon.react
                tag="span"
                className="pachyderm-mount-icon-padding"
              />
            </button>
          </div>
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
                className="pachyderm-button-link"
                onClick={() => setShouldShowAddressInput(true)}
              >
                Change Address
              </button>
            </>
          ) : (
            <div
              style={{
                width: '100%',
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <div
                style={{
                  width: '100%',
                  display: 'flex',
                  justifyContent: 'space-between',
                  marginBottom: '1rem',
                }}
              >
                <span className="pachyderm-mount-config-subheading">
                  {authConfig.pachd_address
                    ? 'Update Configuration'
                    : 'Connect To a Cluster'}
                </span>

                {authConfig.pachd_address && (
                  <button
                    data-testid="Config__pachdAddressCancel"
                    className="pachyderm-button-link"
                    disabled={loading}
                    onClick={() => {
                      setErrorMessage('');
                      setAddressField('');
                      setServerCa('');
                      setShouldShowAddressInput(false);
                      setShowAdvancedOptions(false);
                    }}
                  >
                    <closeIcon.react tag="span" />
                  </button>
                )}
              </div>

              <label
                htmlFor="pachd"
                className="pachyderm-mount-config-address-label"
              >
                Cluster Address
              </label>
              <input
                className="pachyderm-input"
                data-testid="Config__pachdAddressInput"
                name="pachd"
                value={addressField}
                onInput={(e: any) => {
                  if (errorMessage) {
                    setErrorMessage('');
                  }
                  setAddressField(e.target.value);
                }}
                disabled={loading}
                placeholder="grpcs://example.pachyderm.com:30650"
              ></input>

              {showAdvancedOptions && (
                <div className="pachyderm-mount-config-advanced-settings">
                  <label
                    htmlFor="pachd"
                    className="pachyderm-mount-config-address-label"
                    style={{display: 'flex'}}
                  >
                    Server CAs
                    <div
                      className="pachyderm-mount-list-item-status-icon"
                      title="Optional, include if you manage your own certificates."
                    >
                      <infoIcon.react tag="span" />
                    </div>
                  </label>
                  <textarea
                    data-testid="Config__serverCaInput"
                    style={{maxHeight: '200px'}}
                    className="pachyderm-input"
                    value={serverCa}
                    onChange={(e: any) => {
                      setServerCa(e.target.value);
                    }}
                    disabled={loading}
                  ></textarea>
                </div>
              )}

              <div style={{paddingTop: '1rem'}}>
                {loading && (
                  <div
                    className="pachyderm-mount-list-item-status-icon"
                    style={{position: 'static'}}
                  >
                    <LoadingDots />
                  </div>
                )}
                <span className="pachyderm-mount-config-address-error">
                  {errorMessage}
                </span>
              </div>

              <div className="pachyderm-mount-config-advanced-settings-button">
                <button
                  data-testid="Config__advancedSettingsToggle"
                  className="pachyderm-button-link"
                  onClick={() => {
                    if (showAdvancedOptions) {
                      setServerCa('');
                    }
                    setShowAdvancedOptions(!showAdvancedOptions);
                  }}
                  disabled={loading}
                >
                  {showAdvancedOptions
                    ? 'Clear Advanced Settings'
                    : 'Use Advanced Settings'}
                </button>

                <button
                  data-testid="Config__pachdAddressSubmit"
                  className="pachyderm-button pachyderm-mount-config-set-address"
                  style={{width: '100px'}}
                  disabled={loading || !addressField}
                  onClick={updatePachdAddress}
                >
                  Set Address
                </button>
              </div>
            </div>
          )}
        </div>
        {shouldShowLogin && !shouldShowAddressInput && (
          <div className="pachyderm-mount-login-container">
            {reposStatus === 200 ? (
              <button
                data-testid="Config__logout"
                className="pachyderm-button"
                style={{width: '100px'}}
                disabled={loading}
                onClick={callLogout}
              >
                Logout
              </button>
            ) : (
              <button
                data-testid="Config__login"
                className="pachyderm-button"
                style={{width: '100px'}}
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
          <KubernetesElephant width="230px" height="230px" />
        </div>
      </div>
    </>
  );
};

export default Config;
