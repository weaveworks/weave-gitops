import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";
import PolicyDetails from "../../components/Policies/PolicyDetails/PolicyDetails";
import { useGetPolicyDetails } from "../../hooks/Policies";
import { V2Routes } from "../../lib/types";

type Props = {
  className?: string;
  clusterName?: string;
  id: string;
};

const PolicyDetailsPage = ({ className, clusterName, id }: Props) => {
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
      <PolicyDetails policy={data?.policy || {}} />
    </Page>
  );
};

export default styled(PolicyDetailsPage).attrs({
  className: PolicyDetails.name,
})``;
