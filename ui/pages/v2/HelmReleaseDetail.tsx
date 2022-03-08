import * as React from "react";
import styled from "styled-components";
import Heading from "../../components/Heading";
import Page from "../../components/Page";
import ReconciledObjectsTable from "../../components/ReconciledObjectsTable";
import { useGetHelmRelease } from "../../hooks/automations";
import { AutomationKind } from "../../lib/api/core/types.pb";
import { WeGONamespace } from "../../lib/types";

type Props = {
  name: string;
  className?: string;
};

const Info = styled.div`
  padding-bottom: 32px;
`;

function HelmReleaseDetail({ className, name }: Props) {
  const { data, isLoading, error } = useGetHelmRelease(name);
  const helmRelease = data?.helmRelease;

  return (
    <Page loading={isLoading} error={error} className={className}>
      <Info>
        <Heading level={1}>{helmRelease?.name}</Heading>
        <Heading level={2}>{helmRelease?.namespace}</Heading>
      </Info>
      <ReconciledObjectsTable
        kinds={helmRelease?.inventory}
        automationName={helmRelease?.name}
        namespace={WeGONamespace}
        automationKind={AutomationKind.HelmReleaseAutomation}
      />
    </Page>
  );
}

export default styled(HelmReleaseDetail).attrs({
  className: HelmReleaseDetail.name,
})``;
