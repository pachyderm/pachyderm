/* eslint-disable no-constant-condition */
import {Page} from 'playwright';
import {expect} from '@playwright/test';
import {Journey, JourneyObject} from './types';
import {wait} from './utils';

export const loadLandingPage = async (page: Page) => {
  await page.goto('./');

  await expect(page).toHaveTitle(
    /Projects - (HPE ML Data Management|Pachyderm Console)/,
  );

  await expect(page.getByRole('heading', {name: /projects/i})).toHaveText(
    'Projects',
  );
};

const idleLandingPage: Journey = async (page: Page) => {
  await page.goto('./');

  await expect(page).toHaveTitle(
    /Projects - [HPE ML Data Management|Pachyderm Console]/,
  );
  await wait(60 * 1000); // one minute
};

const idleLogsPage: Journey = async (page: Page) => {
  await page.goto(
    './lineage/perf-project-0/pipelines/perf-pipeline-small-files/logs',
  );

  page.getByText('Log Retrieval Limitation'); // Logs are very old and might not exist

  await wait(60 * 1000); // one minute
};

const pageThroughDatums: Journey = async (page: Page) => {
  await page.goto(
    './lineage/perf-project-0/pipelines/perf-pipeline-small-files/logs',
  );

  await page.getByTestId('JobList__listItem').first().click();

  const pageForward = page
    .getByTestId('DatumList__list')
    .getByTestId('Pager__forward');
  const pageBack = page
    .getByTestId('DatumList__list')
    .getByTestId('Pager__backward');
  const MAX_PAGES = 10;

  for (let i = 0; i < MAX_PAGES; i++) {
    await pageForward.click();
    page.getByText('DatumList__listItem'); // Wait for datums to load
  }

  for (let i = 0; i < MAX_PAGES; i++) {
    await pageBack.click();
    page.getByText('DatumList__listItem'); // Wait for datums to load
  }
};

const pageThroughFiles: Journey = async (page: Page) => {
  await page.goto('./lineage/perf-project-0/repos/perf-repo-1/latest');
  const pageForward = page
    .getByTestId('Pager__pager')
    .getByTestId('Pager__forward');
  const pageBack = page
    .getByTestId('Pager__pager')
    .getByTestId('Pager__backward');
  const MAX_PAGES = 10;

  for (let i = 0; i < MAX_PAGES; i++) {
    await pageForward.click();
  }

  for (let i = 0; i < MAX_PAGES; i++) {
    await pageBack.click();
  }
};

// project should have more than 15 repos
const pageThroughRepos: Journey = async (page: Page) => {
  await page.goto('./project/perf-project-1/repos');

  const pageForward = page
    .getByTestId('ReposTable__table')
    .getByTestId('Pager__forward');
  const pageBack = page
    .getByTestId('ReposTable__table')
    .getByTestId('Pager__backward');
  const MAX_PAGES = 8;

  for (let i = 0; i < MAX_PAGES; i++) {
    await pageForward.click();
    page.getByText('RepoListRow__row'); // Wait for repos to load
  }

  for (let i = 0; i < MAX_PAGES; i++) {
    await pageBack.click();
    page.getByText('RepoListRow__row'); // Wait for repos to load
  }
};

// project should have more than 15 pipelines
const pageThroughPipelines: Journey = async (page: Page) => {
  await page.goto('./project/perf-project-1/pipelines');

  const pageForward = page
    .getByTestId('PipelineStepsTable__table')
    .getByTestId('Pager__forward');
  const pageBack = page
    .getByTestId('PipelineStepsTable__table')
    .getByTestId('Pager__backward');
  const MAX_PAGES = 8;

  for (let i = 0; i < MAX_PAGES; i++) {
    await pageForward.click();
    page.getByText('PipelineListRow__row'); // Wait for pipelines to load
  }

  for (let i = 0; i < MAX_PAGES; i++) {
    await pageBack.click();
    page.getByText('PipelineListRow__row'); // Wait for pipelines to load
  }
};

const idleOnJobs: Journey = async (page: Page) => {
  await page.goto('./project/perf-project-0/jobs');

  page.getByText('Log Retrieval Limitation');

  await wait(60 * 1000); // one minute
};

const idleOnSubjobs: Journey = async (page: Page) => {
  await page.goto('./project/perf-project-0/jobs');
  await page.getByTestId('RunsList__row').first().click();
  await page
    .getByRole('tab', {
      name: /subjobs/i,
    })
    .click();
  await wait(60 * 1000); // one minute
};
const idleOnRuntimes: Journey = async (page: Page) => {
  await page.goto('./project/perf-project-0/jobs');
  await page.getByTestId('RunsList__row').first().click();
  await page
    .getByRole('tab', {
      name: /runtimes/i,
    })
    .click();
  await wait(60 * 1000); // one minute
};
const idleOnRepos: Journey = async (page: Page) => {
  await page.goto('./project/perf-project-0/repos/repos');
  await wait(60 * 1000); // one minute
};

// TODO: This one might take too long?
const viewCommitsOnAllRepos: Journey = async (page: Page) => {
  await page.goto('./project/perf-project-0/repos/repos');
  await page
    .getByRole('tab', {
      name: /repos/i,
    })
    .click();

  const rowsCount = await page.getByTestId('RepoListRow__row').count();
  for (const index of Array(rowsCount).keys()) {
    await page
      .getByRole('tab', {
        name: /repos/i,
      })
      .click();

    await page.getByTestId('RepoListRow__row').nth(index).click();
    await page
      .getByRole('tab', {
        name: /commits/i,
      })
      .click();
  }
};

const idleOnDagPipelineSidebar: Journey = async (page: Page) => {
  await page.goto(
    './lineage/perf-project-0/pipelines/perf-pipeline-small-files',
  );
  await wait(60 * 1000);
};

const idleOnDagPipelineOutputSidebar: Journey = async (page: Page) => {
  await page.goto('./lineage/perf-project-0/repos/perf-pipeline-small-files');
  await wait(60 * 1000);
};

const checkAllProjectSummaries: Journey = async (page: Page) => {
  await page.goto('./');
  await page.getByRole('button', {name: /view: your projects/i}).click();
  await page.getByRole('menuitem', {name: /all projects/i}).click();
  const projectRows = await page.getByRole('row').count();
  for (const index of Array(projectRows).keys()) {
    await page.getByRole('row').nth(index).click();
  }
};

export const journeys: JourneyObject[] = [
  {name: 'Load the landing page.', journey: loadLandingPage},
  {name: 'Page through datums', journey: pageThroughDatums},
  {name: 'Page through files', journey: pageThroughFiles},
  {name: 'Page through repos', journey: pageThroughRepos},
  {name: 'Page through pipelines', journey: pageThroughPipelines},
  {name: 'View commits on all repos', journey: viewCommitsOnAllRepos},
  {name: 'Idle for one minute on landing page', journey: idleLandingPage},
  {name: 'Idle for one minute on logs page', journey: idleLogsPage},
  {name: 'Idle on jobs page', journey: idleOnJobs},
  {name: 'Idle on subjobs page', journey: idleOnSubjobs},
  {name: 'Idle on runtimes page', journey: idleOnRuntimes},
  {name: 'Idle on repos', journey: idleOnRepos},
  {name: 'Idle on DAG pipeline sidebar', journey: idleOnDagPipelineSidebar},
  {
    name: 'Idle on DAG pipeline output sidebar',
    journey: idleOnDagPipelineOutputSidebar,
  },
  {
    name: 'Click through projects on the landing page',
    journey: checkAllProjectSummaries,
  },
];

export const loginWithMockAdmin: Journey = async (page: Page) => {
  await page.goto('./');
  await page.getByRole('textbox', {name: 'username'}).type('admin');
  await page.getByLabel('Password').type('password');

  await page.getByRole('button', {name: /login/i}).click();
  await page.close();
};
