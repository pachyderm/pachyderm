describe('Download Canvas', () => {
  beforeEach(() => {
    cy.visit('/');
  });

  it("should download a Project's canvas correctly", () => {
    cy.visit('/lineage/3');
    cy.findByRole('button', {name: 'Open DAG controls menu'}).click();
    cy.findByText('Download Canvas').click();

    cy.waitUntil(() => cy.task(
      'readFileMaybe',
      `Solar\ Power\ Data\ Logger\ Team\ Collab.svg`,
    )).should('have.string', 'Solar Power Data Logger Team Collab');
  });
});
