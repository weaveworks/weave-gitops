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
      label: 'Introducing GitOps',
      collapsed: false,
      link: {
        type: 'doc',
        id: 'intro',
      },
      items: [
        {
          type: 'category',
          label: 'Getting Started',
          collapsed: false,
          link: {
            type: 'doc',
            id: 'getting-started/intro',
          },
          items: [
            {
              type: 'category',
              label: '0. Install Weave GitOps',
              collapsed: true,
              link: {
                type: 'doc',
                id: 'installation/index',
              },
              items: [
                'installation/weave-gitops',
                {
                  type: 'category',
                  label: 'Weave GitOps Enterprise',
                  link: {
                    type: 'doc',
                    id: 'installation/weave-gitops-enterprise/index',
                  },
                  items: [
                    'installation/weave-gitops-enterprise/airgap',
                  ],
                },
                'installation/aws-marketplace',
              ],
            },
            'getting-started/ui',
            'getting-started/deploy',
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'Enterprise Edition',
      link: {
        type: 'doc',
        id: 'intro-ee',
      },
      items: [
        'releases',
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
      label: 'Guides',
      items: [
        'guides/setting-up-dex',
        'guides/cert-manager',
        'guides/displaying-custom-metadata',
        'guides/deploying-capa',
        'guides/using-terraform-templates',
        'guides/delivery',
        'guides/flagger-manual-gating',
      ],
    },
    {
      type: 'category',
      label: 'GitOps Run',
      link: {
        type: 'doc',
        id: 'gitops-run/overview',
      },
      items: [
        'gitops-run/get-started',
      ],
    },
    {
      type: 'category',
      label: 'Cluster Management',
      link: {
        type: 'doc',
        id: 'cluster-management/intro',
      },
      items: [
        'cluster-management/getting-started',
        'cluster-management/cluster-api-providers',
        'cluster-management/managing-existing-clusters',
        'cluster-management/provider-identities',
        'cluster-management/deleting-a-cluster',
        'cluster-management/profiles',
        'cluster-management/add-applications',
        'cluster-management/gitrepo-selection',
        'cluster-management/disable-capi',
      ],
    },
    {
      type: 'category',
      label: 'Terraform',
      items: [
        'terraform/overview',
        'terraform/get-started',
        {
          type: 'category',
          label: 'Using Terraform CR',
          items: [
            'terraform/using-terraform-cr/provision-resources-and-auto-approve',
            'terraform/using-terraform-cr/plan-and-manually-apply-terraform-resources',
            'terraform/using-terraform-cr/provision-resources-and-write-output-data',
            'terraform/using-terraform-cr/detect-drifts-only-without-plan-or-apply',
            'terraform/using-terraform-cr/drift-detection-disabled',
            'terraform/using-terraform-cr/set-variables-for-trraform-resources',
            'terraform/using-terraform-cr/custom-backend',
            'terraform/using-terraform-cr/depends-on',
            'terraform/using-terraform-cr/modules',
            'terraform/using-terraform-cr/customize-runner',
          ],
        },
        'terraform/backup-and-restore',
        'terraform/oci-artifact',
        'terraform/aws-eks',
        'terraform/terraform-enterprise',
        'terraform/tfctl',
        'terraform/environment-variables',
      ],
    },
    {
      type: 'category',
      label: 'Pipelines',
      link: {
        type: 'doc',
        id: 'pipelines/intro',
      },
      items: [
        'pipelines/getting-started',
        'pipelines/authorization',
        'pipelines/promoting-applications',
        'pipelines/pipeline-templates',
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
      label: 'Workspaces',
      link: {
        type: 'doc',
        id: 'workspaces/intro',
      },
      items: [
        'workspaces/multi-tenancy',
        'workspaces/view-workspaces',
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
        'policy/weave-policy-profile',
        'policy/configuration',
        'policy/policy-set',
        'policy/policy-configuration',
        'policy/releases',
        'policy/commit-time-checks',
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
        'secrets/bootstraping-secrets',
        'secrets/setup-secrets-operator',
        'secrets/manage-secrets-ui',
        'secrets/self-service',
        'secrets/faq'
      ],
    },
    {
      type: 'category',
      label: 'Templates',
      items: [
        'gitops-templates/templates',
        'gitops-templates/quickstart-templates',
        'gitops-templates/cli',
      ],
    },
    {
      type: 'category',
      label: 'GitOpsSets',
      items: [
        'gitopssets/intro',
        'gitopssets/features',
        'gitopssets/list-generator',
        'gitopssets/git-generator',
        'gitopssets/matrix-generator',
        'gitopssets/pull-request-generator',
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
