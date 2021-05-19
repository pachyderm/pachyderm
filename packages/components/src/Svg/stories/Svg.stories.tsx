import React from 'react';

import {
  AgBiomeLogoSVG,
  AmexSVG,
  ArrowSVG,
  BillingSVG,
  BoxSVG,
  CardsSVG,
  CheckboxCheckedSVG,
  CheckboxSVG,
  CheckmarkSVG,
  ChevronDownSVG,
  ChevronRight,
  CloseSVG,
  CollaborationSVG,
  CopySVG,
  CopyLinkSVG,
  CubeSVG,
  DigitalReasoningLogoSVG,
  DiscoverSVG,
  DocumentationSVG,
  DownloadSVG,
  DrawerElephantSVG,
  DrawerRingsSVG,
  ElephantCtaSVG,
  ElephantHeadExtraLargeSVG,
  ElephantHeadLargeSVG,
  ElephantHeadSVG,
  ElephantSVG,
  EllipsisSVG,
  EnlargeSVG,
  ExclamationErrorSVG,
  ExitSVG,
  GeneralFusionLogoSVG,
  GenericErrorSVG,
  GithubLogoSVG,
  GithubWithTextSVG,
  GoogleLogoSVG,
  GroupSVG,
  InfoSVG,
  KubernetesElephantSVG,
  LinkSVG,
  LogMeInLogoSVG,
  MastercardSVG,
  MaximizeSVG,
  MembersFreeSVG,
  MembersSVG,
  MinusSVG,
  OvalSVG,
  PachydermLogoBaseSVG,
  PachydermLogoFooterSVG,
  PachydermLogoSVG,
  PencilSVG,
  PipelinesSVG,
  PlusSVG,
  ProgressCheckSVG,
  RectangleSVG,
  SearchSVG,
  SectionRectangleSVG,
  SettingsSVG,
  SlackLogoSVG,
  SlackWithTextSVG,
  SupportSVG,
  TimesSVG,
  TrashSVG,
  TriangleSVG,
  VisaSVG,
  WorkspaceSVG,
} from '..';

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
          <AgBiomeLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="DigitalReasoningLogoSVG">
          <DigitalReasoningLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="GeneralFusionLogoSVG">
          <GeneralFusionLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="GithubLogoSVG">
          <GithubLogoSVG width="24" height="24" />
        </SVGWrapper>
        <SVGWrapper title="GithubWithTextSVG">
          <GithubWithTextSVG />
        </SVGWrapper>
        <SVGWrapper title="GoogleLogoSVG">
          <GoogleLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="LogMeInLogoSVG">
          <LogMeInLogoSVG />
        </SVGWrapper>
        <SVGWrapper title="SlackLogoSVG">
          <SlackLogoSVG width="54" height="54" />
        </SVGWrapper>
        <SVGWrapper title="SlackWithTextSVG">
          <SlackWithTextSVG />
        </SVGWrapper>
      </div>
    </div>
  );
};

export const Pachyderm = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        <SVGWrapper title="BoxSVG">
          <BoxSVG />
        </SVGWrapper>
        <SVGWrapper title="CubeSVG">
          <CubeSVG />
        </SVGWrapper>
        <SVGWrapper title="DrawerRingsSVG">
          <DrawerRingsSVG viewBox="0 0 287 173" />
        </SVGWrapper>
        <SVGWrapper title="DrawerElephantSVG">
          <DrawerElephantSVG viewBox="0 0 235 298" />
        </SVGWrapper>
        <SVGWrapper title="ElephantCtaSVG">
          <ElephantCtaSVG viewBox="0 0 500 625" />
        </SVGWrapper>
        <SVGWrapper title="ElephantHeadExtraLargeSVG">
          <ElephantHeadExtraLargeSVG viewBox="0 0 127 118" />
        </SVGWrapper>
        <SVGWrapper title="ElephantHeadLargeSVG">
          <ElephantHeadLargeSVG viewBox="0 0 127 118" />
        </SVGWrapper>
        <SVGWrapper title="ElephantHeadSVG">
          <ElephantHeadSVG />
        </SVGWrapper>
        <SVGWrapper title="ElephantSVG">
          <ElephantSVG width="641.58" height="588.92" />
        </SVGWrapper>
        <SVGWrapper title="GenericErrorSVG">
          <GenericErrorSVG viewBox="0 0 701 251" />
        </SVGWrapper>
        <SVGWrapper title="KubernetesElephantSVG">
          <KubernetesElephantSVG width="583" height="389" />
        </SVGWrapper>
        <SVGWrapper title="MembersFreeSVG">
          <MembersFreeSVG viewBox="0 0 500 339" />
        </SVGWrapper>
        <SVGWrapper title="OvalSVG">
          <OvalSVG />
        </SVGWrapper>
        <SVGWrapper title="PachydermLogoBaseSVG">
          <PachydermLogoBaseSVG width="451" height="97" />
        </SVGWrapper>
        <SVGWrapper title="PachydermLogoFooterSVG">
          <PachydermLogoFooterSVG viewBox="0 0 100 170" />
        </SVGWrapper>
        <SVGWrapper title="PachydermLogoSVG">
          <PachydermLogoSVG viewBox="0 0 160 32" />
        </SVGWrapper>
        <SVGWrapper title="RectangleSVG">
          <RectangleSVG />
        </SVGWrapper>
        <SVGWrapper title="TriangleSVG">
          <TriangleSVG />
        </SVGWrapper>
      </div>
    </div>
  );
};

export const UIIcons = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        <SVGWrapper title="AmexSVG">
          <AmexSVG />
        </SVGWrapper>
        <SVGWrapper title="ArrowSVG">
          <ArrowSVG />
        </SVGWrapper>
        <SVGWrapper title="BillingSVG">
          <BillingSVG />
        </SVGWrapper>
        <SVGWrapper title="CardsSVG">
          <CardsSVG width="121" height="18" />
        </SVGWrapper>
        <SVGWrapper title="CheckboxSVG">
          <CheckboxSVG />
        </SVGWrapper>
        <SVGWrapper title="CheckboxCheckedSVG">
          <CheckboxCheckedSVG />
        </SVGWrapper>
        <SVGWrapper title="CheckmarkSVG">
          <CheckmarkSVG />
        </SVGWrapper>
        <SVGWrapper title="ChevronDownSVG">
          <ChevronDownSVG />
        </SVGWrapper>
        <SVGWrapper title="ChevronRight">
          <ChevronRight />
        </SVGWrapper>
        <SVGWrapper title="CloseSVG">
          <CloseSVG />
        </SVGWrapper>
        <SVGWrapper title="CollaborationSVG">
          <CollaborationSVG />
        </SVGWrapper>
        <SVGWrapper title="CopySVG">
          <CopySVG />
        </SVGWrapper>
        <SVGWrapper title="CopyLinkSVG">
          <CopyLinkSVG />
        </SVGWrapper>
        <SVGWrapper title="DiscoverSVG">
          <DiscoverSVG />
        </SVGWrapper>
        <SVGWrapper title="DocumentationSVG">
          <DocumentationSVG />
        </SVGWrapper>
        <SVGWrapper title="DownloadSVG">
          <DownloadSVG />
        </SVGWrapper>
        <SVGWrapper title="EllipsisSVG">
          <EllipsisSVG />
        </SVGWrapper>
        <SVGWrapper title="EnlargeSVG">
          <EnlargeSVG />
        </SVGWrapper>
        <SVGWrapper title="ExclamationErrorSVG">
          <ExclamationErrorSVG />
        </SVGWrapper>
        <SVGWrapper title="ExitSVG">
          <ExitSVG />
        </SVGWrapper>
        <SVGWrapper title="GroupSVG">
          <GroupSVG />
        </SVGWrapper>
        <SVGWrapper title="InfoSVG">
          <InfoSVG />
        </SVGWrapper>
        <SVGWrapper title="LinkSVG">
          <LinkSVG />
        </SVGWrapper>
        <SVGWrapper title="MastercardSVG">
          <MastercardSVG />
        </SVGWrapper>
        <SVGWrapper title="MaximizeSVG">
          <MaximizeSVG width="17" height="17" />
        </SVGWrapper>
        <SVGWrapper title="MembersSVG">
          <MembersSVG />
        </SVGWrapper>
        <SVGWrapper title="MinusSVG">
          <MinusSVG width="20" height="2" />
        </SVGWrapper>
        <SVGWrapper title="PencilSVG">
          <PencilSVG />
        </SVGWrapper>
        <SVGWrapper title="PipelinesSVG">
          <PipelinesSVG />
        </SVGWrapper>
        <SVGWrapper title="PlusSVG">
          <PlusSVG width="14" height="14" />
        </SVGWrapper>
        <SVGWrapper title="ProgressCheckSVG">
          <ProgressCheckSVG
            viewBox="0 0 18 13"
            className={styles.progressCheck}
          />
        </SVGWrapper>
        <SVGWrapper title="SearchSVG">
          <SearchSVG />
        </SVGWrapper>
        <SVGWrapper title="SectionRectangleSVG">
          <SectionRectangleSVG />
        </SVGWrapper>
        <SVGWrapper title="SettingsSVG">
          <SettingsSVG />
        </SVGWrapper>
        <SVGWrapper title="SupportSVG">
          <SupportSVG />
        </SVGWrapper>
        <SVGWrapper title="TimesSVG">
          <TimesSVG width="16" height="16" />
        </SVGWrapper>
        <SVGWrapper title="TrashSVG">
          <TrashSVG width="18" height="18" />
        </SVGWrapper>
        <SVGWrapper title="VisaSVG">
          <VisaSVG />
        </SVGWrapper>
        <SVGWrapper title="WorkspaceSVG">
          <WorkspaceSVG />
        </SVGWrapper>
      </div>
    </div>
  );
};
