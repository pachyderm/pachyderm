# Provenance

Provenance enables Pachyderm users to go back in time and see the state of
data at a particular moment in the past and possibly use it in root cause
analysis, as well as to improve or eliminate issues. Provenance tracks the
dependency between datasets and determines their origins. Therefore,
provenance answers not only the question about where the data comes from,
but also how the data was transformed along the way. Data scientists need
to have confidence in the information with which they operate. They need
to be able to reproduce the results and sometimes go through the whole
data tranformation process from scratch multiple times, which makes data
provenance one of the most critical aspects of data analysis. If your
computations result in unexpected numbers, the first place to look for clues
is the historical data that gives insights into possible flaws in the
transformation chain or the data itself.

Consider the following example. A bank makes a decision about a mortgage
application for a particular individual based on all the information
about the individual included with the application, such as credit score,
annual income, and so on. This data goes through multiple transformation
steps with numerous dependencies and decisions made along the way. If anyone
who is involved in the process disputes the final decision, historical data
is the first place to look for proof of authenticity, as well for possible
data manipulation. Data provenance enables data scientist to track the
progress from its origin to the final decision and, if needed, analyze
why individual decisions were made. With the adoption of General Data
Protection Regulation (GDPR) compliance requirements, data provenance is
becoming not only a necessity but a requirement for many financial,
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
the data and identify better data analysis processes while the original
dataset remains untouched. Later, the best results of multiple experiments can
be combined together.

<!--- Add an example that describes provenance accross repositories  and  branches -->
