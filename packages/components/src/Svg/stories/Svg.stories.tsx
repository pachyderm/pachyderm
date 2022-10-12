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
        <SVGWrapper title="LockSVG">
          <PachydermSVG.LockSVG />
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
        <SVGWrapper title="PipelineSVG">
          <PachydermSVG.PipelineSVG />
        </SVGWrapper>
        <SVGWrapper title="PipelinesSVG">
          <PachydermSVG.PipelinesSVG />
        </SVGWrapper>
        <SVGWrapper title="RectangleSVG">
          <PachydermSVG.RectangleSVG />
        </SVGWrapper>
        <SVGWrapper title="RepoSVG">
          <PachydermSVG.RepoSVG />
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
        <SVGWrapper title="AddCircleSVG">
          <IconsSVG.AddCircleSVG />
        </SVGWrapper>
        <SVGWrapper title="ArrowCircleDownSVG">
          <IconsSVG.ArrowCircleDownSVG />
        </SVGWrapper>
        <SVGWrapper title="ArrowCircleLeftSVG">
          <IconsSVG.ArrowCircleLeftSVG />
        </SVGWrapper>
        <SVGWrapper title="ArrowCircleRightSVG">
          <IconsSVG.ArrowCircleRightSVG />
        </SVGWrapper>
        <SVGWrapper title="ArrowCircleUpSVG">
          <IconsSVG.ArrowCircleUpSVG />
        </SVGWrapper>
        <SVGWrapper title="ArrowDownSVG">
          <IconsSVG.ArrowDownSVG />
        </SVGWrapper>
        <SVGWrapper title="ArrowLeftSVG">
          <IconsSVG.ArrowLeftSVG />
        </SVGWrapper>
        <SVGWrapper title="ArrowRightSVG">
          <IconsSVG.ArrowRightSVG />
        </SVGWrapper>
        <SVGWrapper title="ArrowUpSVG">
          <IconsSVG.ArrowUpSVG />
        </SVGWrapper>
        <SVGWrapper title="BillingSVG">
          <IconsSVG.BillingSVG />
        </SVGWrapper>
        <SVGWrapper title="BookmarkSVG">
          <IconsSVG.BookmarkSVG />
        </SVGWrapper>
        <SVGWrapper title="CheckboxCheckedSVG">
          <IconsSVG.CheckboxCheckedSVG />
        </SVGWrapper>
        <SVGWrapper title="CheckboxSVG">
          <IconsSVG.CheckboxSVG />
        </SVGWrapper>
        <SVGWrapper title="CheckmarkSVG">
          <IconsSVG.CheckmarkSVG />
        </SVGWrapper>
        <SVGWrapper title="ChevronDoubleLeftSVG">
          <IconsSVG.ChevronDoubleLeftSVG />
        </SVGWrapper>
        <SVGWrapper title="ChevronDoubleRightSVG">
          <IconsSVG.ChevronDoubleRightSVG />
        </SVGWrapper>
        <SVGWrapper title="ChevronDownSVG">
          <IconsSVG.ChevronDownSVG />
        </SVGWrapper>
        <SVGWrapper title="ChevronLeftSVG">
          <IconsSVG.ChevronLeftSVG />
        </SVGWrapper>
        <SVGWrapper title="ChevronRightSVG">
          <IconsSVG.ChevronRightSVG />
        </SVGWrapper>
        <SVGWrapper title="ChevronUpSVG">
          <IconsSVG.ChevronUpSVG />
        </SVGWrapper>
        <SVGWrapper title="CloseCircleSVG">
          <IconsSVG.CloseCircleSVG />
        </SVGWrapper>
        <SVGWrapper title="CloseSVG">
          <IconsSVG.CloseSVG />
        </SVGWrapper>
        <SVGWrapper title="CodeSVG">
          <IconsSVG.CodeSVG />
        </SVGWrapper>
        <SVGWrapper title="CopySVG">
          <IconsSVG.CopySVG />
        </SVGWrapper>
        <SVGWrapper title="DirectionsSVG">
          <IconsSVG.DirectionsSVG />
        </SVGWrapper>
        <SVGWrapper title="DocumentAddSVG">
          <IconsSVG.DocumentAddSVG />
        </SVGWrapper>
        <SVGWrapper title="DocumentSVG">
          <IconsSVG.DocumentSVG />
        </SVGWrapper>
        <SVGWrapper title="DownloadSVG">
          <IconsSVG.DownloadSVG />
        </SVGWrapper>
        <SVGWrapper title="EditSVG">
          <IconsSVG.EditSVG />
        </SVGWrapper>
        <SVGWrapper title="EducationSVG">
          <IconsSVG.EducationSVG />
        </SVGWrapper>
        <SVGWrapper title="ExpandSVG">
          <IconsSVG.ExpandSVG />
        </SVGWrapper>
        <SVGWrapper title="ExternalLinkSVG">
          <IconsSVG.ExternalLinkSVG />
        </SVGWrapper>
        <SVGWrapper title="FilterSVG">
          <IconsSVG.FilterSVG />
        </SVGWrapper>
        <SVGWrapper title="FlipSVG">
          <IconsSVG.FlipSVG />
        </SVGWrapper>
        <SVGWrapper title="FullscreenSVG">
          <IconsSVG.FullscreenSVG />
        </SVGWrapper>
        <SVGWrapper title="GlobalIdSVG">
          <IconsSVG.GlobalIdSVG />
        </SVGWrapper>
        <SVGWrapper title="HamburgerSVG">
          <IconsSVG.HomeSVG />
        </SVGWrapper>
        <SVGWrapper title="HomeSVG">
          <IconsSVG.HomeSVG />
        </SVGWrapper>
        <SVGWrapper title="InfoSVG">
          <IconsSVG.InfoSVG />
        </SVGWrapper>
        <SVGWrapper title="JobsSVG">
          <IconsSVG.JobsSVG />
        </SVGWrapper>
        <SVGWrapper title="LinkSVG">
          <IconsSVG.LinkSVG />
        </SVGWrapper>
        <SVGWrapper title="MinimizeSVG">
          <IconsSVG.MinimizeSVG />
        </SVGWrapper>
        <SVGWrapper title="NotebooksSVG">
          <IconsSVG.NotebooksSVG />
        </SVGWrapper>
        <SVGWrapper title="OverflowSVG">
          <IconsSVG.OverflowSVG />
        </SVGWrapper>
        <SVGWrapper title="PanelLeftSVG">
          <IconsSVG.PanelLeftSVG />
        </SVGWrapper>
        <SVGWrapper title="PanelRightSVG">
          <IconsSVG.PanelRightSVG />
        </SVGWrapper>
        <SVGWrapper title="RotateSVG">
          <IconsSVG.RotateSVG />
        </SVGWrapper>
        <SVGWrapper title="SearchSVG">
          <IconsSVG.SearchSVG />
        </SVGWrapper>
        <SVGWrapper title="SettingsSVG">
          <IconsSVG.SettingsSVG />
        </SVGWrapper>
        <SVGWrapper title="ShrinkSVG">
          <IconsSVG.ShrinkSVG />
        </SVGWrapper>
        <SVGWrapper title="SpinnerSVG">
          <IconsSVG.SpinnerSVG />
        </SVGWrapper>
        <SVGWrapper title="SupportSVG">
          <IconsSVG.SupportSVG />
        </SVGWrapper>
        <SVGWrapper title="TeamsSVG">
          <IconsSVG.TeamsSVG />
        </SVGWrapper>
        <SVGWrapper title="TerminalSVG">
          <IconsSVG.TerminalSVG />
        </SVGWrapper>
        <SVGWrapper title="TrashSVG">
          <IconsSVG.TrashSVG />
        </SVGWrapper>
        <SVGWrapper title="UpdatedCircleSVG">
          <IconsSVG.UpdatedCircleSVG />
        </SVGWrapper>
        <SVGWrapper title="UploadSVG">
          <IconsSVG.UploadSVG />
        </SVGWrapper>
        <SVGWrapper title="UserSVG">
          <IconsSVG.UserSVG />
        </SVGWrapper>
        <SVGWrapper title="ViewIconSVG">
          <IconsSVG.ViewIconSVG />
        </SVGWrapper>
        <SVGWrapper title="ViewListSVG">
          <IconsSVG.ViewListSVG />
        </SVGWrapper>
        <SVGWrapper title="WorkspacesSVG">
          <IconsSVG.WorkspacesSVG />
        </SVGWrapper>
      </div>
    </div>
  );
};

export const FileIcons = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        <SVGWrapper title="FileAudioSVG">
          <IconsSVG.FileAudioSVG />
        </SVGWrapper>
        <SVGWrapper title="FileCSVSVG">
          <IconsSVG.FileCSVSVG />
        </SVGWrapper>
        <SVGWrapper title="FileDocSVG">
          <IconsSVG.FileDocSVG />
        </SVGWrapper>
        <SVGWrapper title="FileFolderSVG">
          <IconsSVG.FileFolderSVG />
        </SVGWrapper>
        <SVGWrapper title="FileImageSVG">
          <IconsSVG.FileImageSVG />
        </SVGWrapper>
        <SVGWrapper title="FileJSONSVG">
          <IconsSVG.FileJSONSVG />
        </SVGWrapper>
        <SVGWrapper title="FileUnknownSVG">
          <IconsSVG.FileUnknownSVG />
        </SVGWrapper>
        <SVGWrapper title="FileVideoSVG">
          <IconsSVG.FileVideoSVG />
        </SVGWrapper>
      </div>
    </div>
  );
};

export const StatusIcons = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        <SVGWrapper title="StatusCheckmarkSVG">
          <IconsSVG.StatusCheckmarkSVG />
        </SVGWrapper>
        <SVGWrapper title="StatusRunningSVG">
          <IconsSVG.StatusRunningSVG />
        </SVGWrapper>
        <SVGWrapper title="StatusPausedSVG">
          <IconsSVG.StatusPausedSVG />
        </SVGWrapper>
        <SVGWrapper title="StatusWarningSVG">
          <IconsSVG.StatusWarningSVG />
        </SVGWrapper>
        <SVGWrapper title="StatusBusySVG">
          <IconsSVG.StatusBusySVG />
        </SVGWrapper>
        <SVGWrapper title="StatusBlockedSVG">
          <IconsSVG.StatusBlockedSVG />
        </SVGWrapper>
      </div>
    </div>
  );
};

export const BillingIcons = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        <SVGWrapper title="AmexSVG">
          <IconsSVG.AmexSVG />
        </SVGWrapper>
        <SVGWrapper title="CardsSVG">
          <IconsSVG.CardsSVG width="121" height="18" />
        </SVGWrapper>
        <SVGWrapper title="DiscoverSVG">
          <IconsSVG.DiscoverSVG />
        </SVGWrapper>
        <SVGWrapper title="MastercardSVG">
          <IconsSVG.MastercardSVG />
        </SVGWrapper>
        <SVGWrapper title="VisaSVG">
          <IconsSVG.VisaSVG />
        </SVGWrapper>
      </div>
    </div>
  );
};
