.. _doc-style-guide:

Pachyderm Technical Documentation Style Guide
=============================================

This document provides main guidelines for creating technical content
that describes Pachyderm concepts and operations. This style guide is
based on Google Developer Documentation Style Guide and serves as a quick
reference for everyone who wants to contribute to the Pachyderm documentation.
For a more detailed overview, see the `Google Developer Documentation
Style Guide <https://developers.google.com/style/>`__

Overview
--------

We welcome all contributions to the Pachyderm technical documentation and are
happy to help incorporate your content into our docs! We hope that this
document will assits in answering some of your questions about our
contributing guidelines.

Writing Style
-------------

Friendly but not overly colloquial. Avoid jargon, idioms, and references
to pop culture. Use shorter sentences and words over longer alternatives.

Things to avoid:

* "Please" and "thank you".
* Exclamation marks.
* Announcement of features that have not yet been developed.
* All caps to emphasize the importance.
* Parenthesis as much as possible. Use commas instead.

.. list-table::
   :widths: 20 20
   :header-rows: 1

   * - Do not use:
     - Use:
   * - We'll walk you through the installation of the product X. It
       might be a bit difficult, but don't worry, we are here to help.
     - This guide walks you through the process of installation of the
       product X.

Write in the present tense and in second person
-----------------------------------------------

Say "you" instead of "we" and use the present tense where possible. Only use
the future tense when talking about the events that will not happen immediately
but sometime in the future. The future tense introduces uncertainty about
when an action takes places. Therefore, in most cases, use the present tense.

.. list-table::
    :widths: 20 20
    :header-rows: 1

    * - Do not use:
      - Use:
    * - We are going to create a new configuration file that will describe
        our deployment.
      - To create a new configuration file that describes our deployment,
        complete the following steps.

Write for an international audience
-----------------------------------

Avoid idioms and jargon and write in simple American English. The content
that you are writing might later be translated into other foreign languages.
Translating simple short phrases is much easier than long sentences. Use
consistent terminology and avoid misplacing modifiers. Spell out abbreviations
on the first occurrence.

.. list-table::
   :widths: 20 20
   :header-rows: 1

   * - Do not use:
     - Use:
   * - After completing these steps, you are off to the races!
     - After you complete these steps, you can start using Pachyderm.

Don't use:  aka LGTM.

Write in active voice
---------------------

Sentences written in active voice are easier for the reader to understand.
A well-written text has about 95% of sentences written in active voice.
Use passive voice only when the performer of the action is unknown or
to avoid blaming the user for an error.

.. list-table::
   :widths: 20 20
   :header-rows: 1

   * - Do not use:
     - Use:
   * - This behavior means that ``transform.err_cmd`` can be used to
       ignore failed datums.
     - You can use ``transform.err_cmd`` to ignore failed datums.

Put the condition before the steps
----------------------------------

If your sentence has a condition, start the sentence with the conditional
clause and add the descriptive instructions after the clause.

.. list-table::
   :widths: 20 20
   :header-rows: 1

   * - Do not use:
     - Use:
   * - See the Spark documentation for more information.
     - For more information, see the Spark documentation.

Use numbered lists for a sequence of steps
------------------------------------------

If the user needs to follow a set of instructions, organize them in a
numbered list rather than in a bulleted list. Options can be described in a
bulleted list. An exception to this rule is when you have just one step.

.. list-table::
   :widths: 20 20
   :header-rows: 1

   * - Do not use:
     - Use:
   * - * Create a configuration and run the following command.
     - 1. Create a configuration file.

       2. Run the following command:

Break your content into smaller chunks
--------------------------------------

Users do not read the whole body of the text. Instead, they skip and
scan through looking for the text structures that stand out, such as
headings, numbered and bulleted lists, tables, and so on. Try to structure
your content so that it is easy to scan through by adding more titles,
organizing instructions in sequences of steps, and adding tables and
lists for properties and descriptions.

Avoid ending a sentence with a preposition
------------------------------------------

Phrasal verbs are a little bit less formal than single-word verbs. If
possible, replace a phrasal word with a single-word verb equivalent and
if you have to use a phrasal word, avoid finishing the sentence with
a preposition.

.. list-table::
   :widths: 20 20
   :header-rows: 1

   * - Do not use:
     - Use:
   * - The `put file` API includes an option for splitting
       **up** the file into separate datums automatically.
     - The `put file` API includes an option for splitting
       the file into separate datums automatically.

Use meaningful links
--------------------

Link text should mean something to the users when they read it. Phrases
like **Click here** and **Read more** do not provide useful information.
They might be good for call-to-action (CA) buttons on the marketing part
of the website, but in technical content they introduce uncertainty and
confusion.

Furthermore, if a user generates a list of links or uses a speech recognition
technology to navigate through the page, they use keywords and phrases,
such as "Click <text>". Generic links are not helpful for them.

Also, use a standard phrase *For more information, see <link>* to
introduce a link.

.. list-table::
   :widths: 20 20
   :header-rows: 1

   * - Do not use:
     - Use:
   * - More information about getting your FREE trial token and
       activating the dashboard can be found
       [here](https://pachyderm.readthedocs.io/en/latest/enterprise/deployment.html#activate-via-the-dashboard).
     - For more information, see
       [Activate your token by using the dashboard](https://pachyderm.readthedocs.io/en/latest/enterprise/deployment.html#activate-via-the-dashboard).

Markdown vs reSTructuredText
----------------------------

The Pachyderm documentation uses both markdown and reSTructuredText to
author documentation. You can use any format you like with markdown being
slightly more preferred. Although reSTructuredText includes a rich set of
features for authoring documentation, markdown is more widely adopted by
various developer's communities and is supported by all major open-source
documentation platforms. Therefore, it appears to be a better choice for
authoring the Pachyderm documentation. However, if you are an avid
reSTructuredText advocate, feel free to use ``.rst``. An the end of the day,
technical content is most important.

For the table of contents at the top level and all the descending levels of the
documentation hierarchy, you have to use reSTructuredText. If you add
a new ``.md`` or ``.rst`` file, you must include it to one of ``toctree``
directives. Otherwise, it does not appear in the rendered ReadTheDocs
documentation.

