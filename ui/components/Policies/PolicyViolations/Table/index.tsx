import React from "react";
import { useListPolicyValidations } from "../../../../hooks/policyViolations";
import { ListPolicyValidationsRequest } from "../../../../lib/api/core/core.pb";
import { formatURL } from "../../../../lib/nav";
import { V2Routes } from "../../../../lib/types";
import DataTable, { Field, filterConfig } from "../../../DataTable";
import Link from "../../../Link";
import RequestStateHandler from "../../../RequestStateHandler";
import Timestamp from "../../../Timestamp";
import Severity from "../../Utils/Severity";
import { useFeatureFlags } from "../../../../hooks/featureflags";
import { Kind } from "../../../../lib/api/core/types.pb";

interface Props {
  req: ListPolicyValidationsRequest;
}

export const PolicyViolationsList = ({ req }: Props) => {
  const { data, error, isLoading } = useListPolicyValidations(req);
  const { isFlagEnabled } = useFeatureFlags();

  let initialFilterState = {
    ...filterConfig(data?.violations, "severity"),
  };
  if (isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")) {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(data?.violations, "clusterName"),
    };
  }
  const fields: Field[] = [
    {
      label: "Message",
      value: ({ message, clusterName, id }) => (
        <Link
          to={formatURL(V2Routes.PolicyViolationDetails, {
            id,
            clusterName,
            name: message,
            kind: req.kind,
          })}
          data-violation-message={message}
        >
          {message}
        </Link>
      ),
      textSearchable: true,
      sortValue: ({ message }) => message,
      maxWidth: 300,
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
    ...(req.kind === Kind.Policy
      ? [
          {
            label: "Application",
            value: ({ namespace, entity }) => `${namespace} / ${entity}`,
            sortValue: ({ namespace, entity }) => `${namespace} / ${entity}`,
          },
        ]
      : []),
    {
      label: "Severity",
      value: ({ severity }) => <Severity severity={severity || ""} />,
      sortValue: ({ severity }) => severity,
    },
    ...(req.kind !== Kind.Policy
      ? [
          {
            label: "Violated Policy",
            value: "name",
            sortValue: ({ name }) => name,
          },
        ]
      : []),
    {
      label: "Violation Time",
      value: ({ createdAt }) => <Timestamp time={createdAt} />,
      defaultSort: true,
      sortValue: ({ createdAt }) => {
        const t = createdAt && new Date(createdAt).getTime();
        return t * -1;
      },
    },
  ];
  return (
    <RequestStateHandler loading={isLoading} error={error}>
      {data?.violations && (
        <DataTable
          filters={initialFilterState}
          rows={data?.violations}
          fields={fields}
        />
      )}
    </RequestStateHandler>
  );
};
