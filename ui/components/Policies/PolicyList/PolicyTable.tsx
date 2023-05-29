import React from "react";
import { FC } from "react";
import { useFeatureFlags } from "../../../hooks/featureflags";
import { formatURL } from "../../../lib/nav";
import { V2Routes } from "../../../lib/types";
import DataTable, { filterConfig } from "../../DataTable";
import Link from "../../Link";
import Text from "../../Text";
import PolicyMode from "../Utilis/PolicyMode";
import Severity from "../Utilis/Severity";
import { Policy } from "../../../lib/api/core/core.pb";
import Timestamp from "../../Timestamp";

interface CustomPolicy extends Policy {
  audit?: string;
  enforce?: string;
}

interface Props {
  policies: CustomPolicy[];
}

export const PolicyTable: FC<Props> = ({ policies }) => {
  const { isFlagEnabled } = useFeatureFlags();

  const mappedPolicies = policies.map((policy) => {
    policy.audit = policy.modes?.includes("audit") ? "audit" : "";
    policy.enforce = policy.modes?.includes("admission") ? "enforce" : "";
    policy.clusterName = isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
      ? policy.clusterName
      : "Default";
    return policy;
  });

  let initialFilterState = {
    ...filterConfig(mappedPolicies, "severity"),
    ...filterConfig(mappedPolicies, "enforce"),
    ...filterConfig(mappedPolicies, "audit"),
  };

  if (
    isFlagEnabled("WEAVE_GITOPS_FEATURE_TENANCY") &&
    isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
  ) {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(mappedPolicies, "tenant"),
    };
  }

  if (isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")) {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(mappedPolicies, "clusterName"),
    };
  }
  return (
    <DataTable
      key={mappedPolicies?.length}
      filters={initialFilterState}
      rows={mappedPolicies}
      fields={[
        {
          label: "Policy Name",
          value: ({ clusterName, id, name }) => (
            <Link
              to={formatURL(V2Routes.PolicyDetailsPage, {
                clusterName,
                id: id,
                name: name,
              })}
              data-policy-name={name}
            >
              <Text capitalize semiBold>
                {name}
              </Text>
            </Link>
          ),
          textSearchable: true,
          sortValue: ({ name }) => name,
          maxWidth: 650,
        },
        {
          label: "Category",
          value: "category",
        },
        {
          label: "Audit",
          value: ({ audit }) => <PolicyMode modeName={audit} />,
        },
        {
          label: "Enforce",
          value: ({ enforce }) => (
            <PolicyMode modeName={enforce ? "admission" : ""} />
          ),
        },
        ...(isFlagEnabled("WEAVE_GITOPS_FEATURE_TENANCY") &&
        isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
          ? [{ label: "Tenant", value: "tenant" }]
          : []),
        {
          label: "Severity",
          value: ({ severity }) => <Severity severity={severity || ""} />,
          sortValue: ({ severity }) => severity,
        },
        ...(isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
          ? [
              {
                label: "Cluster",
                value: "clusterName",
                sortValue: ({ clusterName }) => clusterName,
              },
            ]
          : []),
        {
          label: "Age",
          value: ({ createdAt }) => <Timestamp time={createdAt} />,
          defaultSort: true,
          sortValue: ({ createdAt }) => {
            const t = createdAt && new Date(createdAt).getTime();
            return t * -1;
          },
        },
      ]}
    />
  );
};
