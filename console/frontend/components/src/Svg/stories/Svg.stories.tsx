import React from 'react';

import * as HPESvg from '../HPE';
import * as IconsSVG from '../Icons';
import * as PachydermSVG from '../Pachyderm';

import styles from './SvgStory.module.css';

export default {title: 'SVGs'};

type SVGWrapperProps = {
  children?: React.ReactNode;
  title: string;
};

const SVGWrapper: React.FC<SVGWrapperProps> = ({title, children}) => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.image}>{children}</div>
      <div className={styles.text}>{title}</div>
    </div>
  );
};

export const HPE = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        <SVGWrapper title="LogoHpe">
          <HPESvg.LogoHpe />
        </SVGWrapper>
        <SVGWrapper title="ErrorIcon">
          <HPESvg.ErrorIconSVG />
        </SVGWrapper>
        <SVGWrapper title="LockIcon">
          <HPESvg.LockIconSVG />
        </SVGWrapper>
        <SVGWrapper title="EmptyIcon">
          <HPESvg.EmptyIconSVG />
        </SVGWrapper>
      </div>
    </div>
  );
};

export const Pachyderm = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        <SVGWrapper title="LogoElephant">
          <PachydermSVG.LogoElephant />
        </SVGWrapper>
        <SVGWrapper title="DrawerElephantSVG">
          <PachydermSVG.DrawerElephantSVG viewBox="0 0 235 298" />
        </SVGWrapper>
        <SVGWrapper title="ElephantCtaSVG">
          <PachydermSVG.ElephantCtaSVG viewBox="0 0 500 625" />
        </SVGWrapper>
        <SVGWrapper title="ElephantErrorState">
          <PachydermSVG.ElephantErrorState viewBox="0 0 1219 823" />
        </SVGWrapper>
        <SVGWrapper title="ElephantEmptyState">
          <PachydermSVG.ElephantEmptyState />
        </SVGWrapper>
        <SVGWrapper title="GenericErrorSVG">
          <PachydermSVG.GenericErrorSVG viewBox="0 0 701 251" />
        </SVGWrapper>
        <SVGWrapper title="PachydermLogoBaseSVG">
          <PachydermSVG.PachydermLogoBaseSVG width="451" height="97" />
        </SVGWrapper>
        <SVGWrapper title="PachydermLogoFooterSVG">
          <PachydermSVG.PachydermLogoFooterSVG viewBox="0 0 100 170" />
        </SVGWrapper>
      </div>
    </div>
  );
};

export const UIIcons = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        {Object.keys(IconsSVG).map((icon) => {
          // eslint-disable-next-line @typescript-eslint/ban-ts-comment
          // @ts-ignore
          // eslint-disable-next-line import/namespace
          const Icon = IconsSVG[icon];
          return (
            <SVGWrapper key={icon} title={icon}>
              <Icon />
            </SVGWrapper>
          );
        })}
      </div>
    </div>
  );
};
