describe('Repos', () => {
  before(() => {
    cy.setupProject().visit('/');
  })

  beforeEach(() => {
    cy.findAllByText('View Project').eq(0).click();
  });

  afterEach(() => {
    cy.visit('/')
  })

  after(() => {
    cy.deleteReposAndPipelines().logout();
  })

  it('should allow a user to create a repo', () => {
    cy.findByText('Create Repo', {timeout: 12000}).click();

    cy.findByLabelText('Repo Name', {exact: false}).type("NewRepo")
    cy.findByLabelText('Description', {exact: false}).type("New repo description")
    cy.findByText('Create').click();

    cy.findByText('NewRepo').click();
    cy.findByText('New repo description', {timeout: 15000});
  })

  it('should allow a user to upload an image', () => {
    cy.findByText('NewRepo', {timeout: 12000}).click();
    cy.findByText('Upload Files').click();
    cy.findByText('Upload File');
  
    cy.fixture('AT-AT.png', null).as('file')
    cy.waitUntil(() => cy.findByLabelText('Attach File').should('not.be.disabled'));
    cy.findByLabelText('Attach File').selectFile({
      contents: '@file',
      fileName: 'AT-AT.png',
    })
    cy.findByRole('button', {name: 'Upload'}).click();
    cy.findByTestId('UploadInfo__success');
    cy.findByTestId('FullPageModal__close').click();
    // Needs to wait for commit polling to update
    cy.findAllByText('View Files', {timeout: 20000}).first().click();
  })

  it('should allow a user to view files and see differences between commits', () => {
    cy.findByText('NewRepo', {timeout: 12000}).click();
    cy.findAllByText('View Files').first().click();

    cy.findByText('1 File added');
    cy.findByText('AT-AT.png');
  })

  it('should allow a user to delete a repo', () => {
    cy.findByText('NewRepo', {timeout: 12000}).click();
    cy.findByTestId('DeleteRepoButton__link').click();
    cy.findByText('Delete').click();
    cy.findByText('NewRepo').should('not.exist');
  })
});
