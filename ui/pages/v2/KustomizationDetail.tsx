import * as React from "react";
import styled from "styled-components";
import Flex from "../../components/Flex";
import Heading from "../../components/Heading";
import InfoList from "../../components/InfoList";
import Interval from "../../components/Interval";
import KubeStatusIndicator from "../../components/KubeStatusIndicator";
import Link from "../../components/Link";
import Page from "../../components/Page";
import ReconciledObjectsTable from "../../components/ReconciledObjectsTable";
import { useGetKustomization } from "../../hooks/automations";
import { AutomationKind } from "../../lib/api/core/types.pb";
import { formatURL } from "../../lib/nav";
import { V2Routes, WeGONamespace } from "../../lib/types";

type Props = {
  name: string;
  className?: string;
};

const Info = styled.div`
  padding-bottom: 32px;
`;

function KustomizationDetail({ className, name }: Props) {
  const { data, isLoading, error } = useGetKustomization(name);

  const kustomization = data?.kustomization;

  return (
    <Page loading={isLoading} error={error} className={className}>
      <Info>
        <Heading level={1}>{kustomization?.name}</Heading>
        <Heading level={2}>{kustomization?.namespace}</Heading>
        <InfoList
          items={[
            [
              "Source",
              <Link
                to={formatURL(V2Routes.GitRepo, {
                  name: kustomization?.sourceRef.name,
                })}
              >
                {kustomization?.sourceRef.kind}/{kustomization?.sourceRef.name}
              </Link>,
            ],
            [
              "Status",
              <Flex start>
                <KubeStatusIndicator conditions={kustomization?.conditions} />
                <div>
                  &nbsp; Applied revision {kustomization?.lastAppliedRevision}
                </div>
              </Flex>,
            ],
            ["Cluster", ""],
            ["Path", kustomization?.path],
            ["Interval", <Interval interval={kustomization?.interval} />],
            ["Last Updated At", kustomization?.lastHandledReconciledAt],
          ]}
        />
      </Info>
      <ReconciledObjectsTable
        kinds={kustomization?.inventory}
        automationName={kustomization?.name}
        namespace={WeGONamespace}
        automationKind={AutomationKind.KustomizationAutomation}
      />
    </Page>
  );
}

export default styled(KustomizationDetail).attrs({
  className: KustomizationDetail.name,
})``;
