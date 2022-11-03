import React, {memo} from 'react';

import Description from '../Description';

interface ConnectProps {
  name: string;
  pachdAddress: string;
}

const Connect: React.FC<ConnectProps> = ({name, pachdAddress}) => {
  const connectCommand =
    `echo '{"pachd_address": "${pachdAddress}", "source": 2}' | ` +
    `pachctl config set context "${name}" --overwrite && ` +
    `pachctl config set active-context "${name}"`;

  return (
    <Description
      id="connect"
      copyText={connectCommand}
      title="Run the following in your terminal:"
    >
      {connectCommand}
    </Description>
  );
};

export default memo(Connect);
