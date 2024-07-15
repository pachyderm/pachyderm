import classnames from 'classnames';
import React, {ReactNode} from 'react';

import {
  ElephantEmptyState,
  ElephantErrorState,
  EmptyIconSVG,
  ErrorIconSVG,
} from '@dash-frontend/../components/src';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';

import styles from './BrandedIcon.module.css';

type IconProps = {
  className?: string;
  communityEditionIcon?: ReactNode;
  enterpriseEditionIcon?: ReactNode;
  disableDefaultStyling?: boolean;
};

export const BrandedEmptyIcon: React.FC<IconProps> = ({
  className,
  communityEditionIcon,
  enterpriseEditionIcon,
  disableDefaultStyling = false,
}) => {
  const {enterpriseActive} = useEnterpriseActive();

  return enterpriseActive ? (
    enterpriseEditionIcon ? (
      <>{enterpriseEditionIcon}</>
    ) : (
      <EmptyIconSVG
        className={classnames(className, {
          [styles.enterpriseImage]: !disableDefaultStyling,
        })}
      />
    )
  ) : communityEditionIcon ? (
    <>{communityEditionIcon}</>
  ) : (
    <ElephantEmptyState
      className={classnames(className, {
        [styles.elephantImage]: !disableDefaultStyling,
      })}
    />
  );
};

export const BrandedErrorIcon: React.FC<IconProps> = ({
  className,
  communityEditionIcon,
  enterpriseEditionIcon,
  disableDefaultStyling = false,
}) => {
  const {enterpriseActive} = useEnterpriseActive();

  return enterpriseActive ? (
    enterpriseEditionIcon ? (
      <>{enterpriseEditionIcon}</>
    ) : (
      <ErrorIconSVG
        className={classnames(className, {
          [styles.enterpriseImage]: !disableDefaultStyling,
        })}
      />
    )
  ) : communityEditionIcon ? (
    <>{communityEditionIcon}</>
  ) : (
    <ElephantErrorState
      className={classnames(className, {
        [styles.elephantImage]: !disableDefaultStyling,
      })}
    />
  );
};
