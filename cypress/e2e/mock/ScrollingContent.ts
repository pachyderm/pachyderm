const inspectListItemScrolling = (testId: string, expectedLength: number) => {
  cy.findAllByTestId(testId).should('have.length', expectedLength);
  cy.findAllByTestId(testId).first().should('be.visible');
  cy.findAllByTestId(testId).last().should('not.be.visible');
  cy.findAllByTestId(testId).last().scrollIntoView().should('be.visible');
  cy.isInViewport(() => cy.findAllByTestId(testId).last());
};

describe(
  'Console Scrolling Content',
  {
    viewportHeight: 600,
    viewportWidth: 1000,
  },
  () => {
  beforeEach(() => {
    cy.visit('/')
  });

  // These tests ensure that scrollable content involving long lists of items don't cut off any items at the bottom when scrolling

  it('should display the last item properly when scrolling a list of projects', () => {
    cy.findAllByRole('row').should('have.length', 7);
    cy.findByText('Data Cleaning Process').should('be.visible');
    cy.findByText('Solar Power Data Logger Team Collab').should('be.visible');
    cy.findByText('Solar Panel Data Sorting').should('not.be.visible');
    cy.findByTestId('Landing__view').scrollTo('bottom')
    cy.findByText('Solar Panel Data Sorting').parent().parent().parent().parent().should('be.visible');
    cy.isInViewport(() => cy.findByText('Solar Panel Data Sorting').parent().parent().parent().parent());
  });

  it('should display the last item properly when scrolling a list of project jobs', () => {
    cy.findByText('Solar Power Data Logger Team Collab').click();
    inspectListItemScrolling('JobListItem__job', 9);
  });

  it('should display the last item properly when scrolling a list of project jobs in lineage and list view', () => {
    cy.findAllByText(/^View(\sProject)*$/).eq(1).click();
    cy.findByText('Jobs').click();
    inspectListItemScrolling('JobListItem__job', 9);

    cy.get(`[aria-label="Close"]`).click();
    cy.findByText('View List').click();
    cy.findByText('Jobs').click();
    inspectListItemScrolling('JobListItem__job', 9);
  });

  it('should display the last item properly when scrolling job details in lineage and list view', () => {
    cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
    cy.findByText('Jobs').click();
    cy.findByTestId('JobListItem__job').click();
    cy.findByText('Total Datums').should('be.visible');
    cy.findByText('dataFailed:').should('not.be.visible');
    cy.findByTestId('InfoPanel__description').scrollTo('bottom')
    cy.findByText('dataFailed:').should('be.visible')
    cy.isInViewport(() => cy.findByText('dataFailed:'));

    cy.findByText('View List').click();
    cy.findByText('Total Datums').should('be.visible');
    cy.findByText('dataFailed:').should('not.be.visible');
    cy.findByTestId('InfoPanel__description').scrollTo('bottom')
    cy.findByText('dataFailed:').should('be.visible')
    cy.isInViewport(() => cy.findByText('dataFailed:'));
  });

  it('should display the last item properly when scrolling a list of repos and pipelines in list view', () => {
    cy.findAllByText(/^View(\sProject)*$/).eq(5).scrollIntoView().click();
    cy.findByText('View List').click();
    cy.findByText('Repositories').click();
    inspectListItemScrolling('ListItem__row', 34);

    cy.findByText('Pipelines').click();
    inspectListItemScrolling('ListItem__row', 27);
  });

  it('should display the last item properly when scrolling info from repos in lineage and list view', () => {
    cy.findAllByText(/^View(\sProject)*$/).eq(1).click();
    cy.findByText('View List').click();
    cy.findByText('Info').click();
    cy.get(`[aria-labelledby="info"]`).children().first().children().should('have.length', 8);
    cy.get(`[aria-labelledby="info"]`).children().first().children().first().should('be.visible');
    cy.get(`[aria-labelledby="info"]`).children().first().children().last().should('be.visible');
    cy.isInViewport(() => cy.get(`[aria-labelledby="info"]`).children().first().children().last());

    cy.findByText('View DAG').click();
    cy.get(`[aria-labelledby="info"]`).children().first().children().should('have.length', 6);
    cy.get(`[aria-labelledby="info"]`).children().first().children().first().should('be.visible');
    cy.get(`[aria-labelledby="info"]`).children().first().children().last().should('be.visible');
    cy.isInViewport(() => cy.get(`[aria-labelledby="info"]`).children().first().children().last());
  });

  it('should display the last item properly when scrolling info from pipelines in lineage and list view', () => {
    cy.findAllByText(/^View(\sProject)*$/).eq(1).click();
    cy.findByText('View List').click();
    cy.findByText('Pipelines').click();
    cy.findByText('Pipeline Info').click();
    cy.get(`[aria-labelledby="info"]`).children().first().children().should('have.length', 22);
    cy.get(`[aria-labelledby="info"]`).children().first().children().eq(1).should('be.visible');
    cy.get(`[aria-labelledby="info"]`).children().first().children().last().should('not.be.visible');
    cy.findByTestId("PipelineDetails__scrollableContent").scrollTo('bottom')
    cy.get(`[aria-labelledby="info"]`).children().first().children().last().should('be.visible');
    cy.isInViewport(() => cy.get(`[aria-labelledby="info"]`).children().first().children().last());

    cy.findByText('View DAG').click();
    cy.get(`[aria-labelledby="info"]`).children().first().children().should('have.length', 20);
    cy.get(`[aria-labelledby="info"]`).children().first().children().eq(1).should('be.visible');
    cy.get(`[aria-labelledby="info"]`).children().first().children().last().should('not.be.visible');
    cy.findByTestId("PipelineDetails__scrollableContent").scrollTo('bottom')
    cy.get(`[aria-labelledby="info"]`).children().first().children().last().should('be.visible');
    cy.isInViewport(() => cy.get(`[aria-labelledby="info"]`).children().first().children().last());
  });

  it('should display the last item properly when scrolling job overview from pipelines in lineage and list view', () => {
    cy.findAllByText(/^View(\sProject)*$/).eq(1).click();
    cy.findByText('View List').click();
    cy.findByText('Pipelines').click();
  
    cy.findByText('Most Recent Job ID').should('be.visible');
    cy.findByText('dataTotal:').should('not.be.visible');
    cy.findByTestId("PipelineDetails__scrollableContent").scrollTo('bottom')
    cy.findByText('Most Recent Job ID').should('not.be.visible');
    cy.findByText('dataTotal:').should('be.visible');
    cy.isInViewport(() => cy.findByText('dataTotal:'));

    cy.findByText('View DAG').click();    
    cy.findByText('Most Recent Job ID').should('be.visible');
    cy.findByText('dataTotal:').should('not.be.visible');
    cy.findByTestId("PipelineDetails__scrollableContent").scrollTo('bottom')
    cy.findByText('Most Recent Job ID').should('not.be.visible');
    cy.findByText('dataTotal:').should('be.visible');
    cy.isInViewport(() => cy.findByText('dataTotal:'));
  });

  it('should display the last item properly when scrolling pipeline specs in lineage and list view', () => {
    /* 
    The app seems to require a full loading of `/`. Otherwise it will redirect
    when visiting '/project/1/pipelines/montage'. This was happening when
    Cypress had to do a full page load to skip the tutorial.

    Since that is temporarily disabled, we can get the same behavior by just
    waiting for some page text to render. I think this is happening because the
    app needs to load an auth token and put it in local storage.
    */
    cy.findByText("Projects")
    cy.visit('/project/1/pipelines/montage')

    cy.findByText('Spec').click();
    cy.findByText('v4tech/imagemagick').should('be.visible');
    cy.findByText('priorityClassName:').should('not.be.visible');
    cy.findByTestId("PipelineDetails__scrollableContent").scrollTo('bottom')
    cy.findByText('priorityClassName:').should('be.visible');
    cy.isInViewport(() => cy.findByText('priorityClassName:'));

    cy.findByText('View DAG').click();
    cy.findByText('Spec').click();
    cy.findByText('v4tech/imagemagick').should('be.visible');
    cy.findByText('priorityClassName:').should('not.be.visible');
    cy.findByTestId("PipelineDetails__scrollableContent").scrollTo('bottom')
    cy.findByText('priorityClassName:').should('be.visible');
    cy.isInViewport(() => cy.findByText('priorityClassName:'));
  });
});


