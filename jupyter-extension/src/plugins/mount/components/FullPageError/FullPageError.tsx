import React from 'react';
import {GenericError, StatusWarning} from '../../../../utils/components/Svgs';
import {HealthCheck} from 'plugins/mount/types';

type FullPageErrorProps = {
  healthCheck: HealthCheck;
};

const FullPageError: React.FC<FullPageErrorProps> = ({healthCheck}) => {
  return (
    <div
      className={'pachyderm-mount-base'}
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div>
        <GenericError width="280px" height="130px" />
      </div>

      <div
        style={{
          display: 'flex',
          flexDirection: 'row',
          alignItems: 'center',
          paddingBottom: '1rem',
        }}
      >
        <StatusWarning />
        <span style={{paddingLeft: '.5rem'}}>
          Looks like there was an error
        </span>
      </div>
      <div data-testid="FullPageError__message">{healthCheck.message}</div>
    </div>
  );
};

export default FullPageError;
