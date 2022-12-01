import React from 'react';

import * as CompaniesSVG from '../Companies';
import * as IconsSVG from '../Icons';
import * as PachydermSVG from '../Pachyderm';

import styles from './SvgStory.module.css';

export default {title: 'SVGs'};

type SVGWrapperProps = {
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

export const Companies = () => {
  return (
    <div className={styles.base}>
      <div className={styles.grid}>
        <SVGWrapper title="AgBiomeLogoSVG">
          <CompaniesSVG.AgBiomeLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="DigitalReasoningLogoSVG">
          <CompaniesSVG.DigitalReasoningLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="GeneralFusionLogoSVG">
          <CompaniesSVG.GeneralFusionLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="GithubLogoSVG">
          <CompaniesSVG.GithubLogoSVG width="24" height="24" />
        </SVGWrapper>
        <SVGWrapper title="GithubWithTextSVG">
          <CompaniesSVG.GithubWithTextSVG />
        </SVGWrapper>
        <SVGWrapper title="GoogleLogoSVG">
          <CompaniesSVG.GoogleLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="LogMeInLogoSVG">
          <CompaniesSVG.LogMeInLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="SlackLogoSVG">
          <CompaniesSVG.SlackLogoSVG width="54" height="54" />
        </SVGWrapper>
        <SVGWrapper title="SlackWithTextSVG">
          <CompaniesSVG.SlackWithTextSVG />
        </SVGWrapper>
      </div>
    </div>
  );
};

export const Pachyderm = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        <SVGWrapper title="BillingTrunkSVG">
          <PachydermSVG.BillingTrunkSVG viewBox="0 0 589 718" />
        </SVGWrapper>
        <SVGWrapper title="BoxSVG">
          <PachydermSVG.BoxSVG />
        </SVGWrapper>
        <SVGWrapper title="CollaborationSVG">
          <PachydermSVG.CollaborationSVG />
        </SVGWrapper>
        <SVGWrapper title="CubeSVG">
          <PachydermSVG.CubeSVG />
        </SVGWrapper>
        <SVGWrapper title="DrawerElephantSVG">
          <PachydermSVG.DrawerElephantSVG viewBox="0 0 235 298" />
        </SVGWrapper>
        <SVGWrapper title="DrawerRingsSVG">
          <PachydermSVG.DrawerRingsSVG viewBox="0 0 287 173" />
        </SVGWrapper>
        <SVGWrapper title="ElephantCtaSVG">
          <PachydermSVG.ElephantCtaSVG viewBox="0 0 500 625" />
        </SVGWrapper>
        <SVGWrapper title="ElephantErrorState">
          <PachydermSVG.ElephantErrorState />
        </SVGWrapper>
        <SVGWrapper title="ElephantEmptyState">
          <PachydermSVG.ElephantEmptyState />
        </SVGWrapper>
        <SVGWrapper title="ElephantHeadExtraLargeSVG">
          <PachydermSVG.ElephantHeadExtraLargeSVG viewBox="0 0 127 118" />
        </SVGWrapper>
        <SVGWrapper title="ElephantHeadLargeSVG">
          <PachydermSVG.ElephantHeadLargeSVG viewBox="0 0 127 118" />
        </SVGWrapper>
        <SVGWrapper title="ElephantHeadSVG">
          <PachydermSVG.ElephantHeadSVG />
        </SVGWrapper>
        <SVGWrapper title="ElephantSVG">
          <PachydermSVG.ElephantSVG width="641.58" height="588.92" />
        </SVGWrapper>
        <SVGWrapper title="GenericErrorSVG">
          <PachydermSVG.GenericErrorSVG viewBox="0 0 701 251" />
        </SVGWrapper>
        <SVGWrapper title="KubernetesElephantSVG">
          <PachydermSVG.KubernetesElephantSVG width="583" height="389" />
        </SVGWrapper>
        <SVGWrapper title="MembersFreeSVG">
          <PachydermSVG.MembersFreeSVG viewBox="0 0 500 339" />
        </SVGWrapper>
        <SVGWrapper title="OvalSVG">
          <PachydermSVG.OvalSVG />
        </SVGWrapper>
        <SVGWrapper title="PachydermLogoBaseSVG">
          <PachydermSVG.PachydermLogoBaseSVG width="451" height="97" />
        </SVGWrapper>
        <SVGWrapper title="PachydermLogoFooterSVG">
          <PachydermSVG.PachydermLogoFooterSVG viewBox="0 0 100 170" />
        </SVGWrapper>
        <SVGWrapper title="PachydermLogoSVG">
          <PachydermSVG.PachydermLogoSVG viewBox="0 0 160 32" />
        </SVGWrapper>
        <SVGWrapper title="PipelinesSVG">
          <PachydermSVG.PipelinesSVG />
        </SVGWrapper>
        <SVGWrapper title="RectangleSVG">
          <PachydermSVG.RectangleSVG />
        </SVGWrapper>
        <SVGWrapper title="SectionRectangleSVG">
          <PachydermSVG.SectionRectangleSVG />
        </SVGWrapper>
        <SVGWrapper title="TriangleSVG">
          <PachydermSVG.TriangleSVG />
        </SVGWrapper>
        <SVGWrapper title="WorkspaceSVG">
          <PachydermSVG.WorkspaceSVG />
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
