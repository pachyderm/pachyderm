describe('Download Canvas', () => {
  beforeEach(() => {
    cy.visit('/');
  });

  it("should download a Project's canvas correctly", () => {
    /* 
    The app seems to require a full loading of `/`. Otherwise it will redirect
    when visiting '/project/1/pipelines/montage'. This was happening when
    Cypress had to do a full page load to skip the tutorial.

    Since that is temporarily disabled, we can get the same behavior by just
    waiting for some page text to render. I think this is happening because the
    app needs to load an auth token and put it in local storage.
    */
    cy.findByText('Projects');
    cy.visit('/lineage/Solar-Power-Data-Logger-Team-Collab');

    cy.findByRole('button', {name: 'Open DAG controls menu'}).click();
    cy.findByText('Download Canvas').click();

    cy.waitUntil(() =>
      cy.task('readFileMaybe', `Solar-Power-Data-Logger-Team-Collab.svg`),
    ).should('have.string', 'Solar-Power-Data-Logger-Team-Collab');
  });
});
