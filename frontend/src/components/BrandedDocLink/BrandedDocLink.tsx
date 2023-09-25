import React from 'react';

import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import {Link, ExternalLinkSVG, Icon} from '@pachyderm/components';

import styles from './BrandedDocLink.module.css';

type BrandedDocLinkProps = {
  pathWithoutDomain: string;
};

const BrandedDocLink: React.FC<
  BrandedDocLinkProps & React.ComponentPropsWithoutRef<typeof Link>
> = ({pathWithoutDomain, children, ...rest}) => {
  const {enterpriseActive} = useEnterpriseActive();

  const domain = enterpriseActive ? 'mldm.pachyderm.com' : 'docs.pachyderm.com';
  const version = 'latest';
  const pathNoLeadingSlash = pathWithoutDomain.replace(/^\//, '');
  const to = `https://${domain}/${version}/${pathNoLeadingSlash}`;

  return (
    <Link externalLink to={to} {...rest}>
      <span className={styles.linkText}>
        {children}{' '}
        <Icon className={styles.externalIcon} color="inherit" small>
          <ExternalLinkSVG aria-hidden />
        </Icon>
      </span>
    </Link>
  );
};

export default BrandedDocLink;
