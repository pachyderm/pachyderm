new Crawler({
  appId: "5ZDILAAVOQ",
  apiKey: "",
  rateLimit: 8,
  startUrls: [
    "https://docs.pachyderm.com/1.13.x/",
    "https://docs.pachyderm.com/2.0.x/",
    "https://docs.pachyderm.com/2.1.x/",
    "https://docs.pachyderm.com/latest/",
    "https://docs.pachyderm.com/2.3.x/",
  ],
  renderJavaScript: false,
  sitemaps: [
    "https://docs.pachyderm.com/1.13.x/sitemap.xml",
    "https://docs.pachyderm.com/2.0.x/sitemap.xml",
    "https://docs.pachyderm.com/2.1.x/sitemap.xml",
    "https://docs.pachyderm.com/latest/sitemap.xml",
    "https://docs.pachyderm.com/2.3.x/sitemap.xml",
  ],
  exclusionPatterns: [],
  ignoreCanonicalTo: false,
  discoveryPatterns: ["https://docs.pachyderm.com/**"],
  schedule: "at 00:00 on Friday",
  actions: [
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.1.x/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.1.x"],
            },
            pageRank: "5",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/1.13.x/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["1.13.x"],
            },
            pageRank: "5",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.0.x/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.0.x"],
            },
            pageRank: "5",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/latest/concepts/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["latest"],
            },
            pageRank: "10",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/latest/how-tos/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["latest"],
            },
            pageRank: "5",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/latest/getting-started/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["latest"],
            },
            pageRank: "10",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/latest/deploy-manage/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["latest"],
            },
            pageRank: "10",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/latest/reference/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["latest"],
            },
            pageRank: "5",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/latest/enterprise/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["latest"],
            },
            pageRank: "5",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/latest/troubleshooting/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["latest"],
            },
            pageRank: "1",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/latest/contributing/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["latest"],
            },
            pageRank: "1",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.3.x/concepts/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.3.x"],
            },
            pageRank: "10",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.3.x/how-tos/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.3.x"],
            },
            pageRank: "5",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.3.x/getting-started/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.3.x"],
            },
            pageRank: "10",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.3.x/deploy-manage/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.3.x"],
            },
            pageRank: "10",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.3.x/reference/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.3.x"],
            },
            pageRank: "5",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.3.x/enterprise/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.3.x"],
            },
            pageRank: "5",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.3.x/troubleshooting/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.3.x"],
            },
            pageRank: "1",
          },
          indexHeadings: true,
        });
      },
    },
    {
      indexName: "pachyderm",
      pathsToMatch: ["https://docs.pachyderm.com/2.3.x/contributing/**"],
      recordExtractor: ({ $, helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: "article h1",
            content:
              "article h3, article h4, article h5, article p, article li",
            lvl0: {
              selectors: ".md-tabs__link.md-tabs__link--active",
              defaultValue: "Documentation",
            },
            lvl2: "article h2",
            version: {
              defaultValue: ["2.3.x"],
            },
            pageRank: "1",
          },
          indexHeadings: true,
        });
      },
    },
  ],
  initialIndexSettings: {
    pachyderm: {
      attributesForFaceting: ["type", "lang", "version"],
      attributesToRetrieve: ["hierarchy", "content", "anchor", "url"],
      attributesToHighlight: ["hierarchy", "hierarchy_camel", "content"],
      attributesToSnippet: ["content:10"],
      camelCaseAttributes: ["hierarchy", "hierarchy_radio", "content"],
      searchableAttributes: [
        "unordered(hierarchy_radio_camel.lvl0)",
        "unordered(hierarchy_radio.lvl0)",
        "unordered(hierarchy_radio_camel.lvl1)",
        "unordered(hierarchy_radio.lvl1)",
        "unordered(hierarchy_radio_camel.lvl2)",
        "unordered(hierarchy_radio.lvl2)",
        "unordered(hierarchy_radio_camel.lvl3)",
        "unordered(hierarchy_radio.lvl3)",
        "unordered(hierarchy_radio_camel.lvl4)",
        "unordered(hierarchy_radio.lvl4)",
        "unordered(hierarchy_radio_camel.lvl5)",
        "unordered(hierarchy_radio.lvl5)",
        "unordered(hierarchy_radio_camel.lvl6)",
        "unordered(hierarchy_radio.lvl6)",
        "unordered(hierarchy_camel.lvl0)",
        "unordered(hierarchy.lvl0)",
        "unordered(hierarchy_camel.lvl1)",
        "unordered(hierarchy.lvl1)",
        "unordered(hierarchy_camel.lvl2)",
        "unordered(hierarchy.lvl2)",
        "unordered(hierarchy_camel.lvl3)",
        "unordered(hierarchy.lvl3)",
        "unordered(hierarchy_camel.lvl4)",
        "unordered(hierarchy.lvl4)",
        "unordered(hierarchy_camel.lvl5)",
        "unordered(hierarchy.lvl5)",
        "unordered(hierarchy_camel.lvl6)",
        "unordered(hierarchy.lvl6)",
        "content",
      ],
      distinct: true,
      attributeForDistinct: "url",
      customRanking: [
        "desc(weight.pageRank)",
        "desc(weight.level)",
        "asc(weight.position)",
      ],
      ranking: [
        "words",
        "filters",
        "typo",
        "attribute",
        "proximity",
        "exact",
        "custom",
      ],
      highlightPreTag: '<span class="algolia-docsearch-suggestion--highlight">',
      highlightPostTag: "</span>",
      minWordSizefor1Typo: 3,
      minWordSizefor2Typos: 7,
      allowTyposOnNumericTokens: false,
      minProximity: 1,
      ignorePlurals: true,
      advancedSyntax: true,
      attributeCriteriaComputedByMinProximity: true,
      removeWordsIfNoResults: "allOptional",
      separatorsToIndex: "_",
    },
  },
});
