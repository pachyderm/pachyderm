describe('Project List', () => {
  before(() => {
    cy.setupProject('error-opencv').visit('/');
  })

  beforeEach(() => {
    cy.findByText('Skip tutorial').click();
    cy.findAllByText(/^View(\sProject)*$/, {timeout: 6000}).eq(0).click();
    cy.findByText('View List', {timeout: 12000}).click();
  });

  afterEach(() => {
    cy.visit('/')
  })

  after(() => {
    cy.deleteReposAndPipelines();
  })

  it('should show the correct number of commits', () => {
    cy.findByText('images').click();
    cy.findByText('Commits').click();
    cy.findAllByTestId('CommitBrowser__commit').should('have.length', 1);
    cy.findByTestId('CommitBrowser__autoCommits').click({force: true});
    cy.findAllByTestId('CommitBrowser__commit').should('have.length', 1);
    cy.findAllByText('edges').eq(0).click();
    cy.findByText('Commits').click();
    cy.findAllByTestId('CommitBrowser__commit').should('have.length', 0);
    cy.findByTestId('CommitBrowser__autoCommits').click({force: true});
    cy.findAllByTestId('CommitBrowser__commit').should('have.length', 1);
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
    const jobsButton = cy.findByTestId("ProjectSideNav__seeJobs");
    jobsButton.click();

    const jobs = cy.findAllByTestId('JobListItem__job', {timeout: 8000});
    jobs.should('have.length', 2);
    jobs.last().click();

    cy.findByTestId('CommitIdCopy__id').invoke('text').then((jobId) => {
      cy.findAllByTestId('JobListItem__job').eq(0).click();
      cy.findByTestId('CommitIdCopy__id').should('not.equal', jobId);
      cy.findByText('Filter by Global ID').click();
      cy.findByTestId('GlobalFilter__name').type(jobId);
      cy.findByText('Apply').click();
      cy.findByTestId('CommitIdCopy__id').should('include.text', jobId);
    });

    cy.findAllByTestId('JobListItem__job').should('have.length', 1);
  });

  it('should apply a global ID filter to repos and redirect', () => {
    cy.findAllByTestId('ListItem__row').should('have.length', 3);
    cy.findByText('edges').click();
    cy.findByTestId('Title__name').should('have.text', 'edges');
    cy.findByText('Commits').click();
  
    cy.findAllByTestId('CommitIdCopy__id').last().invoke('text').then((jobId) => {
      cy.findAllByText('montage').eq(0).click();
      cy.findByTestId('Title__name').should('have.text', 'montage');
      cy.findByText('Filter by Global ID').click();
      cy.findByTestId('GlobalFilter__name').type(jobId);
      cy.findByText('Apply').click();
    });

    cy.findAllByTestId('ListItem__row', {timeout: 8000}).should('have.length', 2);
    cy.findByTestId('Title__name').should('have.text', 'images');
  });

  it('should apply a global ID filter to pipelines and redirect', () => {
    const jobsButton = cy.findByTestId("ProjectSideNav__seeJobs");
    jobsButton.click();

    const jobs = cy.findAllByTestId('JobListItem__job', {timeout: 8000});
    jobs.should('have.length', 2);
    jobs.last().click();

    cy.findByTestId('CommitIdCopy__id').invoke('text').then((jobId) => {
      cy.findByRole('link', {name: 'Pipelines'}).click();
      cy.findByTestId('Title__name').should('have.text', 'montage');
      cy.findAllByTestId('ListItem__row').should('have.length', 2);
      cy.findByText('Filter by Global ID').click();
      cy.findByTestId('GlobalFilter__name').type(jobId);
      cy.findByText('Apply').click();
    });

    cy.findAllByTestId('ListItem__row', {timeout: 8000}).should('have.length', 1);
    cy.findByTestId('Title__name').should('have.text', 'edges');
  });
});
