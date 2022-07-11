# Documentation style guide

This style guide provides editorial guidelines for writing clear and consistent Pachyderm product documentation.

## Voice & tone

Pachyderm's voice in documentation should consistently convey a personality that is friendly, knowledgeable, and empathetic.

The tone of voice may vary depending on the type of content being written. For example, a danger notice may use an urgent and serious tone while a tutorial may use an energetic and instructive tone. Make sure the tone of your documentation aligns with the content. If you aren't sure, ask yourself: "What is the reader likely feeling when presented this content? Why are they here?" and align your language to the most appropriate tone.

---

## Language & grammar

### Use active voice

Write in active voice to enforce clarity and simplicity.

| Don't | Do |
|---|---|
| The update was failing due to a configuration issue. | The update failed due to a configuration issue. |
| A POST request is sent and a response is returned. | Send a POST request; the endpoint sends a response. |

You can break this rule to emphasize an object (the *image* is installed) or de-emphasize a subject (5 errors *were found* in this article).

### Use Global English

Write documentation using Global English. Global English makes comprehension easier for all audiences by avoiding regional idioms/expressions and standardizing spelling words using the US English variant.

### Put conditional clauses first

Order conditional clasues first when drafting sentences; this empowers the reader to skip to the next step when the condition does not apply.

| Don't | Do |
|---|---|
| See this page for more information on how to use this feature. | For more information, see How to use Console. |
| Enable the CustomField setting if you want to map custom fields. | To map custom fields, enable the CustomField setting. |

### Write accessibly

Be mindful of the verbs and adjectives you use to describe software behavior; in particular, avoid abelist langauge.

| Don't | Do |
|---|---|
| TBD | TBD |

---

## Formatting & punctuation

### Code blocks

Use ` ``` ` to wrap long code samples into a code block.

```
This is a code block.
```

### Headers

Use sentence casing for all header titles.

| Don't | Do |
|---|---|
| How to Develop Sentient AI | How to develop sentient AI |
| How to Use The Pachctl CLI | How to use the Pachctl CLI |

### Links

Use meaningful link descriptions, such as the original article's title or a one-line summarization of its contents.

###### Examples:

>
    - See the [Pachyderm Technical Documentation Style Guide](docs-style-guide.md)
    - Use the [official Pachyderm style guide](docs-style-guide.md).

### Lists

- Use **numbered lists** for sequential items, such as instructions.
- Use **unbulleted lists** for all other list types (like this list).

### Commas

Use the serial or Oxford comma in a list of three or more items to minimize any chance of confusion.

| Don't | Do |
|---|---|
| I like swimming, biking and singing. |I like swimming, biking, and singing. |
| I only trust my parents, Madonna and Shakira. | I only trust my parents, Madonna, and Shakira. |

### UI Elements

Bolt UI elements when mentioned in a set of instructions (a numbered list).

> 1. Navigate to **Settings** > **Desktop**.
> 2. Scroll to **Push Notifications**.

---

## Organization

### Use nested headers

Remember to use all header sizes (h1-h6) to organize information. Each header should be a subset of topics or examples more narrow in scope than its parents. This enables readers to both gain more context and mentally parse the information at a glance.

### Publish modular topics

Avoid mixing objectives or use cases in one article; instead, organize and separate your content so that it is task-based. If there are many substasks (such as in a long-form tutorial), organize your content so that each major step is itself an article.

###### Example:

The below outline is 4 articles, with the parent article linking to each modular step.

- How to manage login credentials via the API
  - GET user's ID from `/api/users`
  - POST password change to  `/api/user/credentials`
  - DELETE user's ID from `/api/users`

---

## Images & diagrams