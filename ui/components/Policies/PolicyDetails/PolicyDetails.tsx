import * as React from "react";
import remarkGfm from "remark-gfm";
import styled from "styled-components";
import { useFeatureFlags } from "../../../hooks/featureflags";
import { Policy } from "../../../lib/api/core/core.pb";
import ClusterDashboardLink from "../../ClusterDashboardLink";
import Flex from "../../Flex";
import Text from "../../Text";
import YamlView from "../../YamlView";
import HeaderRows, { Header } from "../Utils/HeaderRows";
import Parameters from "../Utils/Parameters";
import PolicyMode from "../Utils/PolicyMode";
import { ChipWrap, Editor, SectionWrapper } from "../Utils/PolicyUtils";
import Severity from "../Utils/Severity";

type Props = {
  policy: Policy;
};

function PolicyDetails({ policy }: Props) {
  const {
    id,
    tenant,
    tags,
    severity,
    category,
    modes,
    targets,
    clusterName,
    description,
    code,
    howToSolve,
    parameters,
  } = policy;
  const { isFlagEnabled } = useFeatureFlags();
  const defaultHeaders: Header[] = [
    {
      rowkey: "Policy ID",
      value: id,
    },
    {
      rowkey: "Cluster",
      children: <ClusterDashboardLink clusterName={clusterName || ""} />,
      visible: isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER"),
    },
    {
      rowkey: "Tenant",
      value: tenant,
      visible:
        isFlagEnabled("WEAVE_GITOPS_FEATURE_TENANCY") &&
        isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER"),
    },
    {
      rowkey: "Tags",
      children: (
        <Flex id="policy-details-header-tags" gap="4">
          {!!tags && tags?.length > 0 ? (
            tags?.map((tag) => <ChipWrap key={tag} label={tag} />)
          ) : (
            <Text>There is no tags for this policy</Text>
          )}
        </Flex>
      ),
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
      rowkey: "Mode",
      children: modes?.length
        ? modes.map((mode, index) => (
            <PolicyMode key={index} modeName={mode} showName />
          ))
        : "",
    },
    {
      rowkey: "Targeted K8s Kind",
      children: (
        <Flex id="policy-details-header-kinds" gap="4">
          {targets?.kinds?.length ? (
            targets?.kinds?.map((kind) => <ChipWrap key={kind} label={kind} />)
          ) : (
            <Text>There is no kinds for this policy</Text>
          )}
        </Flex>
      ),
    },
  ];

  return (
    <Flex wide tall column gap="32">
      <HeaderRows headers={defaultHeaders} />
      <SectionWrapper title="Description:">
        <Editor children={description || ""} />
      </SectionWrapper>
      <SectionWrapper title="How to solve:">
        <Editor children={howToSolve || ""} remarkPlugins={[remarkGfm]} />
      </SectionWrapper>
      <SectionWrapper title="Policy Code:">
        <YamlView type="rego" yaml={code} />
      </SectionWrapper>
      <SectionWrapper title=" Parameters Values:">
        <Parameters parameters={parameters || []} parameterType="Policy" />
      </SectionWrapper>
    </Flex>
  );
}

export default styled(PolicyDetails).attrs({
  className: PolicyDetails.name,
})``;
