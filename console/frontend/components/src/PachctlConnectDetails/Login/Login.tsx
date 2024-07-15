import React, {memo} from 'react';

import Description from '../Description';

type LoginProps = {
  hasOTP: boolean;
};

const Login: React.FC<LoginProps> = ({hasOTP}) => {
  const loginCommand = hasOTP
    ? 'pachctl auth login --one-time-password'
    : 'pachctl auth login';

  return (
    <Description
      id="login"
      copyText={loginCommand}
      title="Login with:"
      noteText={
        hasOTP ? undefined : 'This opens a new browser tab for authentication.'
      }
    >
      {loginCommand}
    </Description>
  );
};

export default memo(Login);
