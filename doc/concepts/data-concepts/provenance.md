# Provenance

Provenance enables Pachyderm users to go back in time and see the state of
data at a particular moment in the past and possibly use it in root cause
analysis, as well as to improve their code. Provenance tracks the
dependency between datasets and determines their origins. Therefore,
provenance answers not only the question about where the data comes from,
but also how the data was transformed along the way. Data scientists need
to have confidence in the information with which they operate. They need
to be able to reproduce the results and sometimes go through the whole
data transformation process from scratch multiple times, which makes data
provenance one of the most critical aspects of data analysis. If your
computations result in unexpected numbers, the first place to look for clues
is the historical data that gives insights into possible flaws in the
transformation chain or the data itself.

Consider the following example. When a bank makes a decision about a mortgage
application, many factors are taken into consideration, including the credit
credit history, annual income, and so on. This data goes through multiple
automated steps of analysis with numerous dependencies and decisions made
along the way. If the final decision does not satisfy the applicant,
historical data is the first place to look for proof of authenticity,
as well for possible prejudice against the applicant. Data provenance
enables data scientists to track the
progress from its origin to the final decision and make appropriate
changes that address the issue. With the adoption of General Data
Protection Regulation (GDPR) compliance requirements, monitoring data lineage
is becoming a necessity for many financial,
biopharmaceutical, and other organizations that work with sensitive data.

Pachyderm implements provenance for both commits and repositories.
Therefore, you can track not only revisions within one branch but also
understand the connection between the data stored in one repository
with the data that was used to calculate the result in the other
repository; therefore, tracking the data transformation process across
multiple datasets.

Collaboration takes data provenance even further. You can make any dataset
available to other members of your team. When many data scientists have
access to the same data set, they can conduct their own experiments with
the data and identify better data analysis processes in separate branches.



<!--- Add an example that describes provenance accross repositories  and  branches -->
