const versions = require("./versions.json");
/** @type {import('@docusaurus/types').DocusaurusConfig} */
module.exports = {
  title: "Weave GitOps",
  tagline: "The Flux expansion pack from the founders of Flux",
  url: process.env.DOC_URL || "https://docs.gitops.weaveworks.org",
  baseUrl: process.env.DOC_BASE_URL || "/",
  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/favicon_150px.png",
  organizationName: "weaveworks", // Usually your GitHub org/user name.
  projectName: "weave-gitops", // Usually your repo name.
  trailingSlash: true,
  future: {
    v4: {
      removeLegacyPostBuildHeadAttribute: true,
    },
  },
  plugins: [
    () => ({
      // Load yaml files as blobs
      name: "yaml-files-as-blobs",
      configureWebpack: function () {
        return {
          module: {
            rules: [
              {
                test: /\.yaml$/,
                use: [
                  {
                    loader: "file-loader",
                    options: { name: "assets/files/[name]-[hash].[ext]" },
                  },
                ],
              },
            ],
          },
        };
      },
    }),
    [
      '@docusaurus/plugin-client-redirects',
      {
        fromExtensions: ['html', 'htm'], // /myPage.html -> /myPage
        redirects: [
          {
            to: '/docs/intro-weave-gitops/',
            from: ['/docs/getting-started'],
          },
          {
            to: '/docs/intro-weave-gitops/',
            from: '/docs/0.6.2/getting-started'
          },
          {
            to: '/docs/intro-weave-gitops/',
            from: '/docs/releases'
          },
          {
            to: '/docs/enterprise/getting-started/releases-enterprise/',
            from: [
              '/docs/enterprise/releases/',
              '/docs/enterprise/intro',
            ]
          },
          {
            to: '/docs/intro-weave-gitops/',
            from: '/docs/intro'
          },
          {
            to: '/docs/enterprise/getting-started/install-enterprise/',
            from: '/docs/installation/weave-gitops-enterprise/'
          },
          {
            to: '/docs/open-source/getting-started/install-OSS/',
            from: '/docs/installation/weave-gitops/'
          },
          {
            to: '/docs/open-source/getting-started/aws-marketplace/',
            from: '/docs/installation/aws-marketplace/'
          },
        ],
      },
    ],
  ],
  themeConfig: {
    navbar: {
      title: "Weave GitOps",
      logo: {
        alt: "Weave GitOps Logo",
        src: "img/weave-logo.png",
      },
      items: [
        {
          type: "doc",
          docId: "intro-weave-gitops",
          position: "left",
          label: "Docs",
        },
        {
          type: 'docSidebar',
          position: 'left',
          label: 'Reference',
          sidebarId: 'ref',
        },
        {
          to: 'help-and-support',
          label: 'Help & Support',
          position: 'left',
        },
        {
          to: 'feedback-and-telemetry',
          label: 'Feedback & Telemetry',
          position: 'left',
        },
        {
          to: 'security',
          label: 'Security',
          position: 'left',
        },
        {
          href: "https://github.com/weaveworks/weave-gitops",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Support",
          items: [
            {
              label: "Contact Us",
              href: "mailto:info@weaveworks.org",
            },
          ],
        },
        {
          title: "Community",
          items: [
            {
              label: "GitHub",
              href: "https://github.com/weaveworks/weave-gitops",
            },
          ],
        },
        {
          title: 'Follow us',
          items: [
            {
              label: 'Facebook',
              href: 'https://www.facebook.com/WeaveworksInc/',
            },
            {
              label: 'LinkedIn',
              href: 'https://www.linkedin.com/company/weaveworks',
            },
            {
              label: 'Twitter',
              href: 'https://twitter.com/weaveworks',
            },
            {
              label: 'Slack',
              href: 'https://slack.weave.works/',
            },
            {
              label: 'Youtube',
              href: 'https://www.youtube.com/c/WeaveWorksInc',
            },
          ],
        },
      ],
      logo: {
        alt: 'Weaveworks Logo',
        src: 'img/weave-logo.png',
        href: 'https://weaveworks.org',
        width: 35,
        height: 35,
      },
      copyright: `Copyright © ${new Date().getFullYear()} Weaveworks`,
    },
    //   algolia: {
    //     appId: "Z1KEXCDHZE",
    //     apiKey: process.env.ALGOLIA_API_KEY,
    //     indexName: "weave",
    //     // Needed to handle the different versions of docs
    //     contextualSearch: true,
    //
    //     // Optional: Algolia search parameters
    //     // searchParameters: {
    //     //   facetFilters: ['type:content']
    //     // },
    //   },
  },
  scripts: [
    {
      src: 'https://kit.fontawesome.com/73855c6ec3.js',
      async: true,
    },
  ],
  presets: [
    [
      "@docusaurus/preset-classic",
      {
        docs: {
          id: 'default',
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl: "https://github.com/weaveworks/weave-gitops/edit/main/website",
          onlyIncludeVersions: (() => {
            if (process.env.STAGING_BUILD === "true") {
              // Build the last 3 versions for staging to speed it up a bit
              return ["current", ...versions.slice(0, 3)];
            }

            // Return undefined which will fall back to the default of all versions
            return undefined;
          })(),
          disableVersioning: process.env.DISABLE_VERSIONING === "true",
        },
        blog: {
          showReadingTime: true,
          editUrl:
            "https://github.com/weaveworks/weave-gitops/edit/main/website/blog",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
        // gtag: {
        //   trackingID: process.env.GA_KEY,
        //   anonymizeIP: true, // Should IPs be anonymized?
        // },
        sitemap: {
          changefreq: 'weekly',
          priority: 0.5,
          ignorePatterns: [
            '/docs/next/**',
            '/docs/archives/**',
            '/docs/*[0-9].*[0-9].*[0-9]/**'
          ],
          filename: 'sitemap.xml',
        },
      },
    ],
  ],
};
