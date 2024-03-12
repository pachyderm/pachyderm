describe('Download Canvas', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.setupProject().visit('/');
  });

  after(() => {
    cy.deleteReposAndPipelines();
    cy.task('deleteDownloadedFile', 'default.svg');
  });

  beforeEach(() => {
    cy.visit('/');
    cy.findByRole('heading', {name: 'default'});
    cy.task('deleteDownloadedFile', 'default.svg');
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
    cy.findByRole('heading', {name: 'default'});
    cy.visit('/lineage/default');

    cy.findByRole('button', {
      name: 'Open DAG controls menu',
      timeout: 15_000,
    }).click();
    cy.findByText('Download Canvas').click();

    cy.waitUntil(() => cy.task('readDownloadedFileMaybe', `default.svg`)).then(
      (svgContent: unknown) => {
        const parser = new DOMParser();
        const svgDoc = parser.parseFromString(
          svgContent as string,
          'image/svg+xml',
        );
        const textElements = svgDoc.querySelectorAll('text');
        expect(textElements.length).to.be.greaterThan(0);

        const svgTextArray = Array.from(textElements).map(
          (el) => el.textContent,
        );
        svgTextArray.pop(); // Last line contains downloaded datetime that we don't need to assert on

        expect(svgTextArray).to.deep.equal([
          'images',
          'Upload Files',
          'edges',
          'Output',
          'Pipeline',
          'Running',
          'Subjob',
          'Success',
          '1',
        ]);
      },
    );
  });
});
