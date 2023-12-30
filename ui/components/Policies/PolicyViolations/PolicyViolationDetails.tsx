import React from "react";
import remarkGfm from "remark-gfm";

import { AppContext } from "../../../contexts/AppContext";
import { useFeatureFlags } from "../../../hooks/featureflags";
import { PolicyValidation } from "../../../lib/api/core/core.pb";
import { Kind } from "../../../lib/api/core/types.pb";
import { formatURL } from "../../../lib/nav";
import { FluxObject } from "../../../lib/objects";
import { V2Routes } from "../../../lib/types";
import ClusterDashboardLink from "../../ClusterDashboardLink";
import Flex from "../../Flex";
import Link from "../../Link";
import Text from "../../Text";
import Timestamp from "../../Timestamp";
import HeaderRows, { RowItem } from "../Utils/HeaderRows";
import { MarkdownEditor } from "../Utils/MarkdownEditor";
import Parameters from "../Utils/Parameters";
import { SectionWrapper } from "../Utils/PolicyUtils";
import Severity from "../Utils/Severity";

interface ViolationDetailsProps {
  violation: PolicyValidation;
  kind: string;
  entityObject: FluxObject;
}
export const ViolationDetails = ({
  violation,
  entityObject,
  kind,
}: ViolationDetailsProps) => {
  const { isFlagEnabled } = useFeatureFlags();
  const { setDetailModal } = React.useContext(AppContext);
  const {
    severity,
    createdAt,
    category,
    howToSolve,
    description,
    entity,
    namespace,
    occurrences,
    name,
    clusterName,
    parameters,
    policyId,
  } = violation || {};
  const items: RowItem[] = [
    {
      rowkey: "Policy Name",
      children: (
        <Link
          to={formatURL(V2Routes.PolicyDetailsPage, {
            id: policyId,
            clusterName,
            name,
          })}
        >
          {name}
        </Link>
      ),
      value: name,
      visible: kind !== Kind.Policy,
    },
    {
      rowkey: "Cluster",
      children: <ClusterDashboardLink clusterName={clusterName} />,
      visible: isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER"),
    },
    {
      rowkey: "Application",
      value: `${namespace}/${entity}`,
      visible: kind !== Kind.Kustomization && kind !== Kind.HelmRelease,
    },
    {
      rowkey: "Violation Time",
      value: <Timestamp time={createdAt} />,
    },
    {
      rowkey: "Severity",
      children: <Severity severity={severity || ""} />,
    },
    {
      rowkey: "Category",
      value: category,
    },
    {
      rowkey: "Violating Entity",
      children: (
        <Text
          pointer
          size="medium"
          color="primary"
          onClick={() => setDetailModal({ object: entityObject })}
        >
          {entityObject.namespace}/{entityObject.name}
        </Text>
      ),
    },
  ];

  return (
    <Flex wide tall column gap="32">
      <HeaderRows items={items} />
      <SectionWrapper title={` Occurrences ( ${occurrences?.length} )`}>
        <ul className="occurrences">
          {occurrences?.map((item) => (
            <li key={item.message}>
              <Text size="medium"> {item.message}</Text>
            </li>
          ))}
        </ul>
      </SectionWrapper>
      <SectionWrapper title="Description:">
        <MarkdownEditor>{description || ""}</MarkdownEditor>
      </SectionWrapper>
      <SectionWrapper title="How to solve:">
        <MarkdownEditor remarkPlugins={[remarkGfm]}>
          {howToSolve || ""}
        </MarkdownEditor>
      </SectionWrapper>
      <SectionWrapper title=" Parameters Values:">
        <Parameters parameters={parameters} />
      </SectionWrapper>
    </Flex>
  );
};
