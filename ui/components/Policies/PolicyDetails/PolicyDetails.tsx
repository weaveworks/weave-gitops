import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import remarkGfm from "remark-gfm";
import styled from "styled-components";
import { useFeatureFlags } from "../../../hooks/featureflags";
import { PolicyObj as Policy } from "../../../lib/api/core/core.pb";
import { Kind } from "../../../lib/api/core/types.pb";
import ClusterDashboardLink from "../../ClusterDashboardLink";
import Flex from "../../Flex";
import SubRouterTabs, { RouterTab } from "../../SubRouterTabs";
import Text from "../../Text";
import YamlView from "../../YamlView";
import { PolicyViolationsList } from "../PolicyViolations/Table";
import HeaderRows, { Header } from "../Utils/HeaderRows";
import { MarkdownEditor } from "../Utils/MarkdownEditor";
import Parameters from "../Utils/Parameters";
import PolicyMode from "../Utils/PolicyMode";
import { ChipWrap, SectionWrapper } from "../Utils/PolicyUtils";
import Severity from "../Utils/Severity";

type Props = {
  policy: Policy;
};

const PolicyDetails = ({ policy }: Props) => {
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
  const { path } = useRouteMatch();

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
    <SubRouterTabs rootPath={`${path}/details`}>
      <RouterTab name="Details" path={`${path}/details`}>
        <Flex wide tall column gap="32">
          <HeaderRows headers={defaultHeaders} />
          <SectionWrapper title="Description:">
            <MarkdownEditor children={description || ""} />
          </SectionWrapper>
          <SectionWrapper title="How to solve:">
            <MarkdownEditor
              children={howToSolve || ""}
              remarkPlugins={[remarkGfm]}
            />
          </SectionWrapper>
          <SectionWrapper title="Policy Code:">
            <YamlView type="rego" yaml={code} />
          </SectionWrapper>
          <SectionWrapper title="Parameters Definition:">
            <Parameters parameters={parameters || []} parameterType="Policy" />
          </SectionWrapper>
        </Flex>
      </RouterTab>
      <RouterTab name="Violations" path={`${path}/violations`}>
        <PolicyViolationsList
          req={{
            policyId: id,
            clusterName,
            kind: Kind.Policy,
          }}
        />
      </RouterTab>
    </SubRouterTabs>
  );
};

export default styled(PolicyDetails).attrs({
  className: PolicyDetails.name,
})``;
