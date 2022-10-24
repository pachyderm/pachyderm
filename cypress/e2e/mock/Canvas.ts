describe('Download Canvas', () => {
  beforeEach(() => {
    cy.visit('/');
    cy.findByText('Skip tutorial').click();
  });

  it("should download a Project's canvas correctly", () => {
    cy.visit('/lineage/3');
    cy.findAllByTestId('DropdownButton__button').should('have.length', 2);
    cy.findAllByTestId('DropdownButton__button').eq(1).click();
    cy.findByText('Download Canvas').click();

    cy.task(
      'readFileMaybe',
      `Solar\ Power\ Data\ Logger\ Team\ Collab.svg`,
    ).should('have.string', 'Solar Power Data Logger Team Collab');
  });
});
