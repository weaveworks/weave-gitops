const versions = require("./versions.json");
/** @type {import('@docusaurus/types').DocusaurusConfig} */
module.exports = {
  title: "Weave GitOps",
  tagline: "Weave GitOps Documentation",
  url: "https://docs.gitops.weave.works",
  baseUrl: "/",
  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/favicon_150px.png",
  organizationName: "weaveworks", // Usually your GitHub org/user name.
  projectName: "weave-gitops", // Usually your repo name.
  trailingSlash: true,
  plugins: [
    () => ({
      // Load yaml files as blobs
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
          docId: "installation",
          position: "left",
          label: "Installation",
        },
        {
          type: "doc",
          docId: "getting-started",
          position: "left",
          label: "Getting Started",
        },
        {
          to: 'security',
          label: 'Security',
          position: 'left',
        },
        {
          type: "docsVersionDropdown",
          position: "right",
          dropdownActiveClassDisabled: true,
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
              href: "mailto:support@weave.works",
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Weaveworks`,
    },
    algolia: {
      appId: "Z1KEXCDHZE",
      apiKey: "c90c5ade2802df8213d6ac50cf3632f4",
      indexName: "weave",
      // Needed to handle the different versions of docs
      contextualSearch: true,

      // Optional: Algolia search parameters
      // searchParameters: {
      //   facetFilters: ['type:content']
      // },
    },
  },
  presets: [
    [
      "@docusaurus/preset-classic",
      {
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          // Please change this to your repo.
          editUrl: "https://github.com/weaveworks/weave-gitops/edit/main/website",
          lastVersion: versions[0],
          versions: {
            current: {
              label: "main",
            },
          },
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          editUrl:
            "https://github.com/weaveworks/weave-gitops/edit/main/website/blog",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
        gtag: {
          // You can also use your "G-" Measurement ID here.
          // Bogus commit to trigger a build
          trackingID: process.env.GA_KEY,
          // Optional fields.
          anonymizeIP: true, // Should IPs be anonymized?
        },
      },
    ],
  ],
};
