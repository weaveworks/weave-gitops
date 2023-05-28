import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";
import { PolicyTable } from "../../components/Policies/PolicyList/PolicyTable";
import { useListListPolicies } from "../../hooks/Policies";

type Props = {
  className?: string;
};

function PoliciesList({ className }: Props) {
  const { data, isLoading, error } = useListListPolicies({});
  return (
    <Page
      error={error || data?.errors}
      loading={isLoading}
      className={className}
      path={[{ label: "Policies" }]}
    >
      {data?.policies && <PolicyTable policies={data.policies} />}
    </Page>
  );
}

export default styled(PoliciesList).attrs({
  className: PoliciesList.name,
})``;
