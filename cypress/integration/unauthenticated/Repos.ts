describe('Repos', () => {
  before(() => {
    cy.setupProject().visit('/');
  })

  beforeEach(() => {
    cy.findByText('Skip tutorial').click();
    cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
  });

  afterEach(() => {
    cy.visit('/')
  })

  after(() => {
    cy.deleteReposAndPipelines();
  })

  it('should allow a user to create a repo', () => {
    cy.findByText('Create Repo', {timeout: 12000}).click();

    cy.findByLabelText('Repo Name', {exact: false, timeout: 12000}).type("NewRepo")
    cy.findByLabelText('Description', {exact: false}).type("New repo description")
    cy.findByText('Create').click();

    cy.findByText('NewRepo').click();
    cy.findByText('New repo description', {timeout: 15000});
  })

  it('should allow a user to cancel an upload', () => {
    cy.findByText('NewRepo', {timeout: 12000}).click();
    cy.findByText('Upload Files').click();
    cy.findByText('Upload Files');
  
    cy.fixture('AT-AT.png', null).as('file')
    cy.waitUntil(() => cy.findByLabelText('Attach Files').should('not.be.disabled'));
    cy.findByLabelText('Attach Files').selectFile([{
      contents: '@file',
      fileName: 'AT-AT.png',
    }]);

    cy.findByText('AT-AT.png').should('exist');
    cy.findByRole('button', {name: 'Upload'}).click();
    cy.findByTestId('FileCard__cancel').click({force: true}); // file card can sometimes be above the fold
    cy.findByText('AT-AT.png').should('not.exist');
  })

  it('should allow a user to upload images', () => {
    cy.findByText('NewRepo', {timeout: 12000}).click();
    cy.findByText('Upload Files').click();
    cy.findByText('Upload Files');
  
    cy.fixture('AT-AT.png', null).as('file1')
    cy.fixture('puppy.png', null).as('file2')

    cy.waitUntil(() => cy.findByLabelText('Attach Files').should('not.be.disabled'));
    cy.findByLabelText('Attach Files').selectFile([{
      contents: '@file1',
      fileName: 'AT-AT.png',
    }, {
      contents: '@file2',
      fileName: 'puppy.png',
    }])
    
    cy.findByRole('button', {name: 'Upload'}).click();
    cy.findByRole('button', {name: 'Done'}).click();
    cy.findAllByText('Commits').click();
    // Needs to wait for commit polling to update
    cy.findAllByText('View Files', {timeout: 30000}).first().click();
  })

  it('should allow a user to view files and see differences between commits', () => {
    cy.findByText('NewRepo', {timeout: 12000}).click();
    cy.findAllByText('Commits').click();
    cy.findAllByText('View Files').first().click();

    cy.findByText('2 Files added');
    cy.findByText('AT-AT.png');
    cy.findByText('puppy.png');
  })


  it('should not append files on file uploads', () => {
    cy.findByText('NewRepo', {timeout: 12000}).click();

    cy.findAllByText('Commits').click();
    cy.findAllByText('View Files').first().click();

    cy.findByText('451.54 kB')
    cy.findByText('80.59 kB');
    cy.findByTestId('FullPageModal__close').click();

    cy.findByText('Upload Files').click();
    cy.findByText('Upload Files');
  
    cy.fixture('AT-AT.png', null).as('file1')

    cy.waitUntil(() => cy.findByLabelText('Attach Files').should('not.be.disabled'));
    cy.findByLabelText('Attach Files').selectFile([{
      contents: '@file1',
      fileName: 'puppy.png',
    }])
    
    cy.findByRole('button', {name: 'Upload'}).click();
    cy.findByRole('button', {name: 'Done'}).click();

    cy.findAllByText('Commits').click();
    cy.findAllByText('View Files', {timeout: 30000}).should('have.length', 2).first().click();

    cy.findAllByText('80.59 kB').should('have.length', 2);
  })

  it('should allow a user to delete a repo', () => {
    cy.findByText('NewRepo', {timeout: 12000}).click();
    cy.findByTestId('DeleteRepoButton__link').click();
    cy.findByText('Delete').click();
    cy.findByText('NewRepo').should('not.exist');
  })
});
