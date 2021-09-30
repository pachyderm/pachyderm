import classNames from 'classnames';
import React from 'react';

import styles from './LoadingPachyderm.module.css';

type LoadingPachydermProps = {
  freeTrial?: boolean;
  isAnimating?: boolean;
};

const LoadingPachyderm: React.FC<LoadingPachydermProps> = ({
  freeTrial,
  isAnimating,
}) => {
  return (
    <div className={styles.logoWrapper}>
      {isAnimating && (
        <div className={styles.bounce}>
          <div className={styles.spin}>
            <svg
              className={classNames(styles.shape, styles.triangle)}
              viewBox="0 0 70 60"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path d="M35 0L69.641 60H0.358982L35 0Z" fill="currentColor" />
            </svg>
          </div>
          <div className={styles.spin}>
            <svg
              className={classNames(styles.shape, styles.square)}
              viewBox="0 0 60 60"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <rect width="60" height="60" fill="currentColor" />
            </svg>
          </div>
          <div className={styles.spin}>
            <svg
              className={classNames(styles.shape, styles.circle)}
              viewBox="0 0 60 60"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <circle cx="30" cy="30" r="30" fill="currentColor" />
            </svg>
          </div>
          <div className={styles.spin}>
            <svg
              className={classNames(styles.shape, styles.pentagon)}
              viewBox="0 0 63 60"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M31.5 0L62.4093 22.4569L50.603 58.7931H12.397L0.590664 22.4569L31.5 0Z"
                fill="currentColor"
              />
            </svg>
          </div>
        </div>
      )}
      <div className={styles.char}>
        {freeTrial && (
          <svg
            className={styles.freeTrial}
            xmlns="http://www.w3.org/2000/svg"
            xmlnsXlink="http://www.w3.org/1999/xlink"
            width="490"
            height="92"
            viewBox="0 0 490 92"
          >
            <defs>
              <filter
                id="a"
                width="107.9%"
                height="151.4%"
                x="-3.9%"
                y="-16%"
                filterUnits="objectBoundingBox"
              >
                <feOffset dy="5" in="SourceAlpha" result="shadowOffsetOuter1" />
                <feGaussianBlur
                  in="shadowOffsetOuter1"
                  result="shadowBlurOuter1"
                  stdDeviation="2"
                />
                <feColorMatrix
                  in="shadowBlurOuter1"
                  values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.1 0"
                />
              </filter>
              <rect id="b" width="470" height="72" x="0" y="0" rx="6" />
            </defs>
            <g fill="none" fillRule="evenodd">
              <g transform="translate(10 3)">
                <use fill="#000" filter="url(#a)" xlinkHref="#b" />
                <use fill="#E9F3F5" xlinkHref="#b" />
              </g>
              <text
                fill="#4C3D6A"
                fontFamily="Montserrat-SemiBold, Montserrat SemiBold"
                fontSize="38"
                fontWeight="500"
                transform="translate(10 3)"
              >
                <tspan x="121.3" y="49">
                  FREE TRIAL
                </tspan>
              </text>
            </g>
          </svg>
        )}
        <div className={styles.body} />
        <div
          className={classNames(styles.head, {
            [styles.isAnimating]: isAnimating,
          })}
        >
          <div
            className={classNames(styles.skull, {
              [styles.isAnimating]: isAnimating,
            })}
          />
          <div className={styles.trunk} />
        </div>
      </div>
    </div>
  );
};

export default LoadingPachyderm;
