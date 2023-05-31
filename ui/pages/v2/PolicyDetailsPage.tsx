import * as React from "react";
import styled from "styled-components";
import Flex from "../../components/Flex";
import Page from "../../components/Page";
import PolicyDetails from "../../components/Policies/PolicyDetails/PolicyDetails";
import Parameters from "../../components/Policies/Utils/Parameters";
import { useGetPolicyDetails } from "../../hooks/Policies";
import { V2Routes } from "../../lib/types";
import { SectionWrapper } from "../../components/Policies/Utils/PolicyUtils";

type Props = {
  className?: string;
  clusterName?: string;
  id: string;
};

function PolicyDetailsPage({ className, clusterName, id }: Props) {
  const { data, isLoading, error } = useGetPolicyDetails({
    clusterName,
    policyName: id,
  });

  return (
    <Page
      error={error || []}
      loading={isLoading}
      className={className}
      path={[
        { label: "Policies", url: V2Routes.Policies },
        { label: data?.policy?.name || "" },
      ]}
    >
      <Flex wide tall column gap="32">
        <PolicyDetails policy={data?.policy || {}} />
        <SectionWrapper title=" Parameters Values:">
          <Parameters
            parameters={data?.policy?.parameters || []}
            parameterType="Policy"
          />
        </SectionWrapper>
      </Flex>
    </Page>
  );
}

export default styled(PolicyDetailsPage).attrs({
  className: PolicyDetails.name,
})``;
