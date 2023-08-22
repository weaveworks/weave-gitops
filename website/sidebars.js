/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

 module.exports = {
  docs: [
    {
      type: 'category',
      label: 'Introducing Weave GitOps',
      collapsed: false,
      link: {
        type: 'doc',
        id: 'intro-weave-gitops',
      },
      items: [
        {
          type: 'category',
          label: 'Weave GitOps Open Source',
          collapsed: false,
          items: [
            'open-source/getting-started/install-OSS',
            'open-source/getting-started/ui-OSS',
            'open-source/getting-started/deploy-OSS',
            'open-source/getting-started/aws-marketplace',
          ],
        },
        {
          type: 'category',
          label: 'Weave GitOps Enterprise',
          link: {
            type: 'doc',
            id: 'enterprise/getting-started/intro-enterprise',
          },
          items: [
            'enterprise/getting-started/install-enterprise',
            'enterprise/getting-started/install-enterprise-airgap',
            'enterprise/getting-started/releases-enterprise',
          ],
        },
        {
            type: 'link',
            label: 'Version Archives',
            href: '/archives',
        },
      ],
    },
    {
      type: 'category',
      label: 'Access Configuration',
      items: [
        'configuration/recommended-rbac-configuration',
        {
          type: 'category',
          label: 'Securing Access to the Dashboard',
          collapsed: false,
          link: {
            type: 'doc',
            id:'configuration/securing-access-to-the-dashboard',
          },
          items: [
            'configuration/oidc-access',
            'configuration/emergency-user',
          ],
        },
        'configuration/service-account-permissions',
        'configuration/user-permissions',
        'configuration/tls',
      ],
    },
    {
      type: 'category',
      label: 'Cluster Management',
      link: {
        type: 'doc',
        id: 'cluster-management/cluster-management-intro',
      },
      items: [
        'cluster-management/managing-clusters-without-capi',
        'cluster-management/deploying-capa-eks',
        'cluster-management/profiles',
        'cluster-management/cluster-management-troubleshooting',
        {
          type: 'category',
          label: 'Advanced Cluster Management',
          items: [
            'cluster-management/advanced-cluster-management-topics/howto-inject-credentials-into-template'
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'Explorer',
      link: {
        type: 'doc',
        id: 'explorer/intro',
      },
      items: [
        'explorer/getting-started',
        'explorer/configuration',
        'explorer/querying',
        'explorer/operations',
      ],
    },
    {
      type: 'category',
      label: 'GitOps Run',
      link: {
        type: 'doc',
        id: 'gitops-run/gitops-run-overview',
      },
      items: [
        'gitops-run/gitops-run-get-started',
      ],
    },
    {
      type: 'category',
      label: 'GitOpsSets',
      items: [
        'gitopssets/gitopssets-intro',
        'gitopssets/gitopssets-installation',
        'gitopssets/templating-from-generators',
        'gitopssets/gitopssets-api-reference',
        'gitopssets/gitopssets-releases'
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/setting-up-dex',
        'guides/displaying-custom-metadata',
        'guides/fluxga-upgrade',
      ],
    },
    {
      type: 'category',
      label: 'Pipelines',
      link: {
        type: 'doc',
        id: 'pipelines/pipelines-intro',
      },
      items: [
        'pipelines/pipelines-getting-started',
        'pipelines/authorization',
        'pipelines/promoting-applications',
        'pipelines/pipelines-templates',
        'pipelines/pipelines-with-jenkins',
        'pipelines/pipelines-with-tekton',
        {
          type: 'category',
          label: 'Reference',
          items: [
            {
              type: 'category',
              label: 'v1alpha1',
              items: [
                'pipelines/spec/v1alpha1/pipeline',
              ],
            },
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'Policy',
      link: {
        type: 'doc',
        id: 'policy/intro',
      },
      items: [
        'policy/getting-started',
        'policy/authorization',
        'policy/policy',
        'policy/weave-policy-profile',
        'policy/policy-set',
        'policy/policy-configuration',
        'policy/releases',
        'policy/commit-time-checks',
      ],
    },
    {
      type: 'category',
      label: 'Progressive Delivery',
      link: {
        type: 'doc',
        id: 'progressive-delivery/progressive-delivery-flagger-install',
      },
      items: [
        'progressive-delivery/flagger-manual-gating',
      ],
    },
    {
      type: 'category',
      label: 'Secrets',
      link: {
        type: 'doc',
        id: 'secrets/intro',
      },
      items: [
        'secrets/intro',
        'secrets/getting-started',
        'secrets/bootstrapping-secrets',
        'secrets/setup-eso',
        'secrets/setup-sops',
        'secrets/manage-secrets-ui',
        // 'secrets/self-service',
        {
          type: 'category',
          label: 'Reference',
          items: [
            {
              type: 'category',
              label: 'v1alpha1',
              items: [
                'secrets/spec/v1alpha1/secretSync',
              ],
            },
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'Templates',
      link: {
        type: 'doc',
        id: 'gitops-templates/intro',
      },
      items: [
        'gitops-templates/quickstart-templates',
        {
          type: 'category',
          label: 'Creating Templates',
          link: {
            type: 'doc',
            id: 'gitops-templates/creating-templates',
          },
          items: [
            'gitops-templates/resource-templates',
            'gitops-templates/repo-rendered-paths',
            'gitops-templates/profiles',
            'gitops-templates/annotations',
            'gitops-templates/params',
            'gitops-templates/supported-langs',
            'gitops-templates/create-cluster-example',
          ],
        },
        'gitops-templates/cli',
        'gitops-templates/versions',
      ],
    },
    {
      type: 'category',
      label: 'Terraform',
      items: [
        'terraform/terraform-intro',
        'terraform/get-started-terraform',
        'terraform/using-terraform-templates',
      ],
    },
    {
      type: 'category',
      label: 'Workspaces',
      link: {
        type: 'doc',
        id: 'workspaces/intro',
      },
      items: [
        'workspaces/multi-tenancy',
        'workspaces/view-workspaces',
      ],
    }
  ],
  ref: [
    {
      type: 'doc',
      label: 'OSS Helm Reference',
      id: 'references/helm-reference',
    },
    {
      type: 'category',
      label: 'CLI Reference',
      items: [
        {
          type: 'autogenerated',
          dirName: 'references/cli-reference',
        },
      ],
    },
  ],
};
