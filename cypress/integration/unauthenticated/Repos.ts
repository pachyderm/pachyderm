describe('Repos', () => {
  before(() => {
    cy.visit('/');
  })

  beforeEach(() => {
    cy.findByText('Skip tutorial').click();
    cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
    cy.findByText('Create Repo', {timeout: 12000}).click();
    cy.findByLabelText('Repo Name', {exact: false, timeout: 12000}).type("TestRepo")
    cy.findByText('Create').click();
  });

  afterEach(() => {
    cy.exec('pachctl delete repo --all --force')
    cy.visit('/')
  })

  it('should allow a user to create a repo', () => {
    cy.findByText('Create Repo', {timeout: 12000}).click();

    cy.findByLabelText('Repo Name', {exact: false, timeout: 12000}).clear().type("NewRepo")
    cy.findByLabelText('Description', {exact: false}).type("New repo description")
    cy.findByText('Create').click();

    cy.findByText('NewRepo').click();
    cy.findByText('New repo description', {timeout: 15000});
  })

  it('should allow a user to cancel an upload', () => {
    cy.findByText('TestRepo', {timeout: 12000}).click();
    cy.waitUntil(() => cy.findByLabelText('Upload Files').should('not.be.disabled'));
    cy.findByLabelText('Upload Files').click();
    cy.findByText('For large file uploads via CTL');
  
    cy.fixture('AT-AT.png', null).as('file')
    cy.waitUntil(() => cy.findByLabelText('Attach Files').should('not.be.disabled'));
    cy.findByLabelText('Attach Files').selectFile([{
      contents: '@file',
      fileName: 'AT-AT.png',
    }], {force: true});

    cy.findByText('AT-AT.png').should('exist');
    cy.findByLabelText('Upload Selected Files').click();
    cy.findByTestId('FileCard__cancel').click({force: true}); // file card can sometimes be above the fold
    cy.findByText('AT-AT.png').should('not.exist');
  })

  it('should allow a user to upload images and not append files, and view differences between commits', () => {
    cy.findByText('TestRepo', {timeout: 12000}).click();
    cy.waitUntil(() => cy.findByLabelText('Upload Files').should('not.be.disabled'));
    cy.findByLabelText('Upload Files').click();
    cy.findByText('For large file uploads via CTL');
  
    cy.fixture('AT-AT.png', null).as('file1')
    cy.fixture('puppy.png', null).as('file2')

    cy.waitUntil(() => cy.findByLabelText('Attach Files').should('not.be.disabled'));
    cy.findByLabelText('Attach Files').selectFile([{
      contents: '@file1',
      fileName: 'AT-AT.png',
    }, {
      contents: '@file2',
      fileName: 'puppy.png',
    }], {force: true});
    
    cy.waitUntil(() => cy.findByLabelText('Upload Selected Files').should('be.visible').click());
    cy.findByLabelText('Commit Selected Files').click();
    cy.findAllByText('Commits').click();

    // Needs to wait for commit polling to update
    cy.findAllByText('View Files', {timeout: 30000}).first().click();

    cy.findByText('2 Files added');
    cy.findByText('451.54 kB')
    cy.findByText('80.59 kB');
    cy.findByText('AT-AT.png');
    cy.findByText('puppy.png');
  
    cy.findByTestId('FullPageModal__close').click({force: true});

    cy.waitUntil(() => cy.findByLabelText('Upload Files').should('not.be.disabled'));
    cy.findByLabelText('Upload Files').click();
    cy.findByText('For large file uploads via CTL');

    cy.waitUntil(() => cy.findByLabelText('Attach Files').should('not.be.disabled'));
    cy.findByLabelText('Attach Files').selectFile([{
      contents: '@file1',
      fileName: 'puppy.png',
    }], {force: true});
    
    cy.waitUntil(() => cy.findByLabelText('Upload Selected Files').should('be.visible').click());
    cy.findByLabelText('Commit Selected Files').click();

    cy.findAllByText('Commits').click();
    cy.findAllByText('View Files', {timeout: 30000}).should('have.length', 2).first().click();
    cy.findAllByText('80.59 kB').should('have.length', 2);
  })

  it('should allow a user to delete a repo', () => {
    cy.findByText('TestRepo', {timeout: 12000}).click();
    cy.findByTestId('DeleteRepoButton__link').click();
    cy.findByText('Delete').click();
    cy.findByText('TestRepo').should('not.exist');
  })
});
