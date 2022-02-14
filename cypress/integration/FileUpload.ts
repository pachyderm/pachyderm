describe('File Upload', () => {
  beforeEach(() => {
    cy.login().then(() => {
      cy.findAllByText('View Project').eq(0).click();
    });
  });

  afterEach(() => {
    cy.logout();
  });

  it('should allow a user to upload an image', () => {
    cy.findByText('images', {timeout: 8000}).click();
    cy.findByText('Upload Files').click();
    cy.findByText('Upload File');

    cy.fixture('AT-AT.png', null).as('file')
    cy.findByLabelText('Attach File').selectFile({
      contents: '@file',
      fileName: 'AT-AT.png',
    })
    cy.findByRole('button', {name: 'Upload'}).click();
    cy.findByTestId('UploadInfo__success');
    cy.findByTestId('FullPageModal__close').click();
    cy.findAllByText('View Files').first().click();
  })
});
