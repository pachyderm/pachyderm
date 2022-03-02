import {requestAPI} from '../../../../handler';
import React, {useEffect, useState} from 'react';
import {KubernetesElephantSVG} from '@pachyderm/components';
import {AuthConfig} from 'plugins/mount/types';
type ConfigProps = {
  showConfig: boolean;
  setShowConfig: (shouldShow: boolean) => void;
  authenticated?: boolean;
};

const Config: React.FC<ConfigProps> = ({
  showConfig,
  setShowConfig,
  authenticated,
}) => {
  const [currentConfig, setCurrentConfig] = useState<AuthConfig | null>();
  const [loading, setLoading] = useState(false);
  const [addressField, setAddressField] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [shouldShowLogin, setShouldShowLogin] = useState(false);

  useEffect(() => {
    if (showConfig) {
      setup();
    }
  }, [showConfig]);

  const setup = async () => {
    // Decide what state to render config when shown
    const response = await requestAPI<AuthConfig>('config', 'GET');
    setShouldShowLogin(response.cluster_status === 'AUTH_ENABLED');
    //TODO: cluster_status === INVALID show error message for invalid config
    //TODO: cluster_status === AUTH_DISABLED hide the login/logout buttons
    setCurrentConfig(response);
  };

  const updatePachdAddress = async () => {
    setLoading(true);
    try {
      const response = await requestAPI<any>('config', 'PUT', {
        pachd_address: addressField,
      });
      setCurrentConfig(response);
      setShouldShowLogin(response.cluster_status === 'AUTH_ENABLED');
      //TODO: cluster_status === INVALID - This case is same as error from calling set config
      //TODO: cluster_status === AUTH_DISABLED - Close Config screen, detect 200 from get repos
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
        window.open(res.auth_url);
        //TODO: Close Config screen, detect authenticated prop change or use hook from comp lib
      }
    } catch (e) {
      setErrorMessage('error logging in.');
      console.log(e);
    }
    setLoading(false);
  };

  const callLogout = async () => {
    setLoading(true);
    try {
      await requestAPI<any>('auth/_logout', 'PUT');
    } catch (e) {
      setErrorMessage('error logging out.');
      console.log(e);
    }
    setLoading(false);
  };

  return (
    <>
      <div className="pachyderm-mount-config-form-base">
        {authenticated && (
          <button
            onClick={() => {
              setShowConfig(false);
            }}
          >
            {'Back'}
          </button>
        )}
        <div className="pachyderm-mount-config-heading">
          Pachyderm
          <span className="pachyderm-mount-config-subheading">
            Mount Extension
          </span>
        </div>
        {currentConfig && currentConfig.pachd_address && shouldShowLogin ? (
          <>
            <label htmlFor="pachd">Cluster Address</label>
            <span>{currentConfig.pachd_address}</span>
            <button onClick={() => setShouldShowLogin(false)}>
              {'Change Address'}
            </button>
          </>
        ) : (
          <>
            <label htmlFor="pachd">Cluster Address</label>
            <input
              name="pachd"
              value={addressField}
              onInput={(e: any) => setAddressField(e.target.value)}
            ></input>
            <button disabled={loading} onClick={updatePachdAddress}>
              {'Set Address'}
            </button>
            <span>{errorMessage}</span>

            {currentConfig && currentConfig.pachd_address && (
              <button
                disabled={loading}
                onClick={() => setShouldShowLogin(true)}
              >
                {'Cancel'}
              </button>
            )}
          </>
        )}
        {shouldShowLogin && (
          <div className="pachyderm-mount-login-container">
            {authenticated ? (
              <button disabled={loading} onClick={callLogout}>
                Logout
              </button>
            ) : (
              <button disabled={loading} onClick={callLogin}>
                Login
              </button>
            )}
          </div>
        )}
      </div>
      <div className="pachyderm-mount-config-graphic-base">
        <KubernetesElephantSVG />
      </div>
    </>
  );
};

export default Config;
