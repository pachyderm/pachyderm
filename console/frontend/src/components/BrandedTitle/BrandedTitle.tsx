import React from 'react';
import {Helmet} from 'react-helmet';

import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';

const BrandedTitle: React.FC<{title: string}> = ({title}) => {
  const {enterpriseActive} = useEnterpriseActive();

  return (
    <Helmet>
      <title>
        {title
          ? `${title} - ${
              enterpriseActive ? 'HPE ML Data Management' : 'Pachyderm Console'
            }`
          : `${
              enterpriseActive ? 'HPE ML Data Management' : 'Pachyderm Console'
            }`}
      </title>
    </Helmet>
  );
};

export default BrandedTitle;
