import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
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
        link: enterpriseActive ? EMAIL_SUPPORT : SLACK_SUPPORT,
      }}
    />
  );
};

export default ErrorStateSupportLink;
