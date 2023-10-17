describe('Download Canvas', () => {
  beforeEach(() => {
    cy.visit('/');
    cy.findByRole('heading', {name: 'Data-Cleaning-Process'});
    cy.task('deleteDownloadedFile', 'Solar-Power-Data-Logger-Team-Collab.svg');
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
    cy.findByRole('heading', {name: 'Data-Cleaning-Process'});
    cy.visit('/lineage/Solar-Power-Data-Logger-Team-Collab');

    cy.findByRole('button', {
      name: 'Open DAG controls menu',
      timeout: 15_000,
    }).click();
    cy.findByText('Download Canvas').click();

    cy.waitUntil(() =>
      cy.task(
        'readDownloadedFileMaybe',
        `Solar-Power-Data-Logger-Team-Collab.svg`,
      ),
    ).then((svgContent: unknown) => {
      const parser = new DOMParser();
      const svgDoc = parser.parseFromString(
        svgContent as string,
        'image/svg+xml',
      );
      const textElements = svgDoc.querySelectorAll('text');
      expect(textElements.length).to.be.greaterThan(0);

      const svgTextArray = Array.from(textElements).map((el) => el.textContent);
      svgTextArray.pop(); // Last line contains downloaded datetime that we don't need to assert on

      expect(svgTextArray).to.deep.equal([
        'cron',
        'processor',
        'Output',
        'Pipeline',
        'Unknown',
        'Subjob',
        'Success',
      ]);
    });
  });
});
