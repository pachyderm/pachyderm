import {EnterpriseState} from '@graphqlTypes';
import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import {useGetEnterpriseInfoQuery} from '@dash-frontend/generated/hooks';

type ErrorStateSupportLinkProps = {
  title: string;
  message?: string;
};

const ErrorStateSupportLink: React.FC<ErrorStateSupportLinkProps> = ({
  title,
  message = null,
}) => {
  const {data} = useGetEnterpriseInfoQuery();
  const active = data?.enterpriseInfo.state === EnterpriseState.ACTIVE;

  return (
    <EmptyState
      error
      title={title}
      message={message || undefined}
      linkToDocs={{
        text: 'If this issue keeps happening, contact our customer team.',
        link: active
          ? 'mailto:support@pachyderm.com'
          : 'https://pachyderm-users.slack.com/archives/C01SMT73Z41',
      }}
    />
  );
};

export default ErrorStateSupportLink;
