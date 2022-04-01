describe('Project List', () => {
  before(() => {
    cy.setupProject('error-opencv').visit('/');
  })

  beforeEach(() => {
    cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
    cy.findByText('View List', {timeout: 8000}).click();
  });

  afterEach(() => {
    cy.visit('/')
    cy.findAllByText(/^View(\sProject)*$/, {timeout: 8000}).eq(0).click();
    cy.findByText('View Lineage', {timeout: 8000}).click();
    cy.visit('/')
  })

  after(() => {
    cy.deleteReposAndPipelines().logout();
  })

  it('should show the correct number of commits', () => {
    cy.findByText('images').click();
    cy.findAllByTestId('CommitBrowser__commit').should('have.length', 3);
    cy.findAllByText('(10 B)', {exact: false}).should('have.length', 3);
    cy.findAllByText('edges').eq(0).click();
    cy.findAllByTestId('CommitBrowser__commit').should('have.length', 2);
  });

  it('should switch between list view and lineage view', () => {
    cy.findByRole('link', {name: 'Repositories'});
    cy.findByRole('link', {name: 'Pipelines'});
    cy.findAllByTestId('ListItem__row').should('have.length', 3);
    cy.findByTestId('Title__name').should('have.text', 'montage');
    
    cy.findByText('View Lineage').click();

    cy.findByText('Jobs');
    cy.findByText('Reset Canvas');
    cy.findByText('Flip Canvas');
    cy.findByText('Center Selections');

    cy.findByText('View List').click();
  });

  it('should apply a global ID filter to jobs and redirect', () => {
    const jobs = cy.findByTestId("ProjectSideNav__seeJobs");
    jobs.click();

    const job = cy.findByTestId('JobList__projectdefault', {timeout: 8000}).findAllByText("Failure");
    job.should('have.length', 2);
    job.last().click();

    cy.findByTestId('CommitIdCopy__id').invoke('text').then((jobId) => {
      cy.findByTestId('JobList__projectdefault').findAllByText("Failure").eq(0).click();
      cy.findByTestId('CommitIdCopy__id').should('not.equal', jobId);
      cy.findByText('Filter by Global ID').click();
      cy.findByTestId('GlobalFilter__name').type(jobId);
      cy.findByText('Apply').click();
      cy.findByTestId('CommitIdCopy__id').should('include.text', jobId);
    });

    cy.findAllByText("Failure").should('have.length', 2);
  });

  it('should apply a global ID filter to repos and redirect', () => {
    cy.findAllByTestId('ListItem__row').should('have.length', 3);
    cy.findByText('edges').click();
    cy.findByTestId('Title__name').should('have.text', 'edges');
  
    cy.findAllByTestId('CommitIdCopy__id').last().invoke('text').then((jobId) => {
      cy.findByText('montage').click();
      cy.findByTestId('Title__name').should('have.text', 'montage');
      cy.findByText('Filter by Global ID').click();
      cy.findByTestId('GlobalFilter__name').type(jobId);
      cy.findByText('Apply').click();
    });

    cy.findAllByTestId('ListItem__row').should('have.length', 2);
    cy.findByTestId('Title__name').should('have.text', 'images');
  });

  it('should apply a global ID filter to pipelines and redirect', () => {
    const jobs = cy.findByTestId("ProjectSideNav__seeJobs");
    jobs.click();

    const job = cy.findAllByText("Failure", {timeout: 8000});
    job.last().click();

    cy.findByTestId('CommitIdCopy__id').invoke('text').then((jobId) => {
      cy.findByRole('link', {name: 'Pipelines'}).click();
      cy.findByTestId('Title__name').should('have.text', 'montage');
      cy.findAllByTestId('ListItem__row').should('have.length', 2);
      cy.findByText('Filter by Global ID').click();
      cy.findByTestId('GlobalFilter__name').type(jobId);
      cy.findByText('Apply').click();
    });

    cy.findAllByTestId('ListItem__row').should('have.length', 1);
    cy.findByTestId('Title__name').should('have.text', 'edges');
  });
});
