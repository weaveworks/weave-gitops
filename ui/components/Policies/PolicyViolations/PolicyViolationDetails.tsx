import React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { PolicyValidation } from "../../../lib/api/core/core.pb";
import Flex from "../../Flex";
import Link from "../../Link";
import Text from "../../Text";
import Timestamp from "../../Timestamp";

import { AppContext } from "../../../contexts/AppContext";
import { useFeatureFlags } from "../../../hooks/featureflags";
import { FluxObject } from "../../../lib/objects";
import ClusterDashboardLink from "../../ClusterDashboardLink";
import HeaderRows, { Header } from "../Utils/HeaderRows";
import Parameters from "../Utils/Parameters";
import Severity from "../Utils/Severity";

const SectionWrapper = ({ title, children }) => {
  return (
    <Flex column wide gap="8" data-testid="occurrences">
      <Text bold color="neutral30">
        {title}
      </Text>
      {children}
    </Flex>
  );
};

interface ViolationDetailsProps {
  violation: PolicyValidation;
  entityUrl: string;
  entityObject: FluxObject;
}
export const ViolationDetails = ({
  violation,
  entityObject,
  entityUrl,
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
  } = violation || {};

  const headers: Header[] = [
    {
      rowkey: "Policy Name",
      value: name,
    },
    {
      rowkey: "Cluster",
      children: <ClusterDashboardLink clusterName={clusterName} />,
      visible: isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER"),
    },
    {
      rowkey: "Application",
      children: (
        <Link to={entityUrl}>
          {namespace}/{entity}
        </Link>
      ),
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
      <HeaderRows headers={headers} />
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
        <ReactMarkdown children={description || ""} className="editor" />
      </SectionWrapper>
      <SectionWrapper title="How to solve:">
        <ReactMarkdown
          children={howToSolve || ""}
          remarkPlugins={[remarkGfm]}
          className="editor"
        />
      </SectionWrapper>
      <SectionWrapper title=" Parameters Values:">
        <Parameters parameters={parameters} />
      </SectionWrapper>
    </Flex>
  );
};
