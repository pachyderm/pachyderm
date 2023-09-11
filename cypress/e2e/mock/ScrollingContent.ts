const inspectListItemScrolling = (testId: string, expectedLength: number) => {
  cy.findAllByTestId(testId).should('have.length', expectedLength);
  cy.findAllByTestId(testId).first().should('be.visible');
  cy.findAllByTestId(testId).last().should('not.be.visible');
  cy.findAllByTestId(testId).last().scrollIntoView();
  cy.findAllByTestId(testId).last().should('be.visible');
  cy.isInViewport(() => cy.findAllByTestId(testId).last());
};

describe(
  'Console Scrolling Content',
  {
    viewportHeight: 500,
    viewportWidth: 1000,
  },
  () => {
    beforeEach(() => {
      cy.visit('/');
    });

    // These tests ensure that scrollable content involving long lists of items don't cut off any items at the bottom when scrolling

    it('should display the last item properly when scrolling a list of projects', () => {
      cy.findAllByRole('row').should('have.length', 11);
      cy.findByText('Data-Cleaning-Process').should('be.visible');
      cy.findByText('Trait-Discovery').should('not.be.visible');
      cy.findByTestId('Landing__view').scrollTo('bottom');
      cy.findByText('Trait-Discovery')
        .parent()
        .parent()
        .parent()
        .parent()
        .should('be.visible');
      cy.isInViewport(() =>
        cy.findByText('Trait-Discovery').parent().parent().parent().parent(),
      );
    });

    it('should display the last item properly when scrolling a list of project jobs', () => {
      cy.contains(
        '[role="row"]',
        /Solar-Power-Data-Logger-Team-Collab/i,
      ).scrollIntoView();
      cy.contains(
        '[role="row"]',
        /Solar-Power-Data-Logger-Team-Collab/i,
      ).click();
      inspectListItemScrolling('JobListItem__job', 9);
    });

    it('should display the last item properly when scrolling a list of job sets', () => {
      cy.findByRole('button', {
        name: /View project Solar-Power-Data-Logger-Team-Collab/i,
      }).scrollIntoView();
      cy.findByRole('button', {
        name: /View project Solar-Power-Data-Logger-Team-Collab/i,
      }).click();
      cy.findByText('Jobs').click();
      inspectListItemScrolling('RunsList__row', 9);
    });

    it('should display the last item properly when scrolling a list of repos', () => {
      cy.findByRole('button', {
        name: /View project Trait-Discovery/i,
      }).scrollIntoView();
      cy.findByRole('button', {
        name: /View project Trait-Discovery/i,
      }).click();
      cy.findByText('Repositories').click();
      inspectListItemScrolling('RepoListRow__row', 34);
    });

    it('should display the last item properly when scrolling a list of pipelines', () => {
      cy.findByRole('button', {
        name: /View project Trait-Discovery/i,
      }).scrollIntoView();
      cy.findByRole('button', {
        name: /View project Trait-Discovery/i,
      }).click();
      cy.findByText('Pipelines').click();
      inspectListItemScrolling('PipelineListRow__row', 27);
    });

    it('should display the last item properly when scrolling info from pipelines in lineage view', () => {
      cy.findByRole('button', {
        name: /View project Solar-Power-Data-Logger-Team-Collab/i,
      }).scrollIntoView();
      cy.findByRole('button', {
        name: /View project Solar-Power-Data-Logger-Team-Collab/i,
      }).click();
      cy.findByText('DAG').click();
      cy.findByText('Pipeline').click({force: true});
      cy.findByText('Pipeline Info').click();
      cy.get(`[aria-labelledby="info"]`)
        .children()
        .first()
        .children()
        .should('have.length', 18);
      cy.get(`[aria-labelledby="info"]`)
        .children()
        .first()
        .children()
        .eq(1)
        .should('be.visible');
      cy.get(`[aria-labelledby="info"]`)
        .children()
        .first()
        .children()
        .last()
        .should('not.be.visible');
      cy.findByTestId('PipelineDetails__scrollableContent').scrollTo('bottom');
      cy.get(`[aria-labelledby="info"]`)
        .children()
        .first()
        .children()
        .last()
        .should('be.visible');
      cy.isInViewport(() =>
        cy.get(`[aria-labelledby="info"]`).children().first().children().last(),
      );
    });

    it('should display the last item properly when scrolling commits from repos in lineage view', () => {
      cy.findByRole('button', {
        name: /View project Solar-Power-Data-Logger-Team-Collab/i,
      }).scrollIntoView();
      cy.findByRole('button', {
        name: /View project Solar-Power-Data-Logger-Team-Collab/i,
      }).click();
      cy.findByText('DAG').click();
      cy.findByText('cron').click();
      cy.findAllByTestId('CommitList__commit').should('have.length', 5);
      cy.findAllByTestId('CommitList__commit').first().should('not.be.visible');
      cy.findAllByTestId('CommitList__commit').last().should('not.be.visible');
      cy.findAllByTestId('CommitList__commit').last().scrollIntoView();
      cy.findAllByTestId('CommitList__commit').should('be.visible');
      cy.isInViewport(() => cy.findAllByTestId('CommitList__commit').last());
    });

    it('should display the last item properly when scrolling pipeline specs in lineage view', () => {
      cy.findByText('Projects');
      cy.visit('/lineage/Solar-Panel-Data-Sorting/pipelines/montage');

      cy.findByText('Spec').click();
      cy.findByTestId('PipelineDetails__scrollableContent').scrollTo('bottom');
      cy.findByText('"priorityClassName"').should('be.visible');
      cy.isInViewport(() => cy.findByText('"priorityClassName"'));
    });
  },
);
