import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';

type ErrorStateSupportLinkProps = {
  title: string;
  message?: string;
};

const ErrorStateSupportLink: React.FC<ErrorStateSupportLinkProps> = ({
  title,
  message = null,
}) => {
  const {enterpriseActive} = useEnterpriseActive();

  return (
    <EmptyState
      error
      title={title}
      message={message || undefined}
      linkExternal={{
        text: 'If this issue keeps happening, contact our customer team.',
        link: enterpriseActive
          ? 'mailto:support@pachyderm.com'
          : 'https://pachyderm-users.slack.com/archives/C01SMT73Z41',
      }}
    />
  );
};

export default ErrorStateSupportLink;
