import React from 'react';

import Description from '../Description';

export type VerifyProps = {
  children?: React.ReactNode;
  command?: string;
};

const Verify: React.FC<VerifyProps> = ({
  command = 'pachctl version',
  children,
}) => {
  return (
    <>
      <Description
        id="verify"
        copyText={command}
        title="Verify that the installation was successful:"
      >
        {command}
      </Description>
      <Description
        asListItem={false}
        id="response"
        title="Expected system response:"
      >
        <table>
          <tbody>
            <tr>
              <th>COMPONENT</th>
              <th>VERSION</th>
            </tr>
            {children}
          </tbody>
        </table>
      </Description>
    </>
  );
};

export default Verify;
