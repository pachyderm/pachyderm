# Documentation Style Guide

Thank you for taking an interest in contributing to Pachyderm's docs! 🐘 📖

This style guide provides editorial guidelines for writing clear and consistent Pachyderm product documentation. See our [contribution guide](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/doc#pachyderm-documentation) for instructions on how to draft and  submit changes.

## Audience

Pachyderm has two main audiences:

- **MLOPs Engineers**: They install and configure Pachyderm to transform data using pipelines.
- **Data Scientists & Data/ML Engineers**: They plan the development of pipelines and consume the outputs of Pachyderm's data processing to feed AI/ML models.

 Be sure to provide links to pre-requisite or contextual materials whenever possible, as everyone's experience level and career journey is different.

---

## Voice & Tone

Pachyderm's voice in documentation should consistently convey a personality that is friendly, knowledgeable, and empathetic.

The tone of voice may vary depending on the type of content being written. For example, a danger notice may use an urgent and serious tone while a tutorial may use an energetic and instructive tone. Make sure the tone of your documentation aligns with the content. If you aren't sure what tone to convey, ask yourself: "What is the reader likely feeling when presented this content? Why are they here?" and adjust your language to the most appropriate tone.

---

## Language & Grammar

The following guidelines are to be followed loosely; use your best judgment when approaching content -- there are always exceptions.

### Use Active Voice

Write in active voice to enforce clarity and simplicity.

| 👎🚫 Don't | 👍✅ Do |
|---|---|
| The update was failing due to a configuration issue. | The update failed due to a configuration issue. |
| A POST request is sent and a response is returned. | Send a POST request; the endpoint sends a response. |

You can break this rule to emphasize an object (the *image* is installed) or de-emphasize a subject (5 errors *were found* in this article).

### Use Global English

Write documentation using Global English. Global English makes comprehension easier for all audiences by avoiding regional idioms/expressions and standardizing spelling words using the US English variant.

### Put Conditional Clauses First

Order conditional clauses first when drafting sentences; this empowers the reader to skip to the next step when the condition does not apply.

| 👎🚫 Don't | 👍✅ Do |
|---|---|
| See this page for more information on how to use this feature. | For more information, see How to Use Console. |
| Enable the CustomField setting if you want to map custom fields. | To map custom fields, enable the CustomField setting. |

### Write Accessibly

Be mindful of how you describe software behavior and users; in particular, avoid ableist language. Use generic "they/them" and "you" when describing users or actors; avoid the use of "obviously", "simply", "easily" -- every reader has a different level of expertise and familiarity with key concepts/tools.

| 👎🚫 Don't | 👍✅ Do |
|---|---|
| To start, simply enter the following command: | To start, enter the following command: |
| Configuring this setting just requires a simple API call. | Make an API call to configure this setting. |
| The results, without this setting enabled, might look crazy. | Your results may be inconsistent or unreliable without this setting enabled. |

---

## Formatting & Punctuation

### Markdown

All documentation is written using Markdown syntax in `.md` files. See this official [Markdown Cheat Sheet](https://www.markdownguide.org/cheat-sheet/) for a quick introduction to the syntax. 

### Code Blocks

Use ` ``` ` to wrap long code samples into a code block.

```
This is a code block.
```

### Headers

Use title casing for all header titles.

- Capitalize the first and last word.
- Capitalize adjectives, adverbs, nouns, pronouns, and subordinate conjuctions.
- Lowercase articles (a, an, the) and coordinating conjunctions (and, but,for, nor, or, so, yet).

| 👎🚫 Don't | 👍✅ Do |
|---|---|
| How to develop sentient ai | How to Develop Sentient AI |
| How to use the pachctl cli | How to Use the Pachctl CLI |

### Links

Use meaningful link descriptions, such as the original article's title or a one-line summarization of its contents.

**Examples:**

>
    - See the [Pachyderm Technical Documentation Style Guide](docs-style-guide.md)
    - Use the [official Pachyderm style guide](docs-style-guide.md).

### Lists

- Use **numbered lists** for sequential items, such as instructions.
- Use **unbulleted lists** for all other list types (like this list).

### Commas

Use the serial or Oxford comma in a list of three or more items to minimize any chance of confusion.

| 👎🚫 Don't | 👍✅ Do |
|---|---|
| I like swimming, biking and singing. |I like swimming, biking, and singing. |
| I only trust my parents, Madonna and Shakira. | I only trust my parents, Madonna, and Shakira. |

### UI Elements

Bold UI elements when mentioned in a set of instructions (a numbered list).

> 1. Navigate to **Settings** > **Desktop**.
> 2. Scroll to **Push Notifications**.

---

## Organization

### Use Nested Headers

Remember to use all header sizes (h1-h6) to organize information. Each header should be a subset of topics or examples more narrow in scope than its parents. This enables readers to both gain more context and mentally parse the information at a glance.

### Publish Modular Topics

Avoid mixing objectives or use cases in one article; instead, organize and separate your content so that it is task-based. If there are many substasks (such as in a long-form tutorial), organize your content so that each major step is itself an article.

**Examples:**

The below outlines are 4 articles, with the parent article linking to each modular sub-topic.

- How to Locally Deploy Pachyderm
    - MacOS Local Deployment Guide
    - Linux Local Deployment Guide
    - Windows Local Deployment Guide

---

## Images & Diagrams

Visualizations are helpful in learning complex workflows and UIs; however, they can expire quickly and take a lot of effort to maintain. Ask yourself the following questions when deciding whether or not to add visualizations:

- Is the user interface complex enough to warrant a screenshot?
- Can a diagram convey this concept more efficiently than words?
- How frequently does this visual need to be updated?

Also, remember to add alt-text to your visualizations for screen readers.

---

Want to make an update to this style guide? Select **Edit me on Github** and leave a suggestion as a pull request!
