import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Flex from "../../components/Flex";
import KubeStatusIndicator from "../../components/KubeStatusIndicator";
import Link from "../../components/Link";
import Page from "../../components/Page";
import ReconciledObjectsTable from "../../components/ReconciledObjectsTable";
import Text from "../../components/Text";
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

const InfoList = styled(
  ({
    items,
    className,
  }: {
    className?: string;
    items: { [key: string]: any };
  }) => {
    return (
      <table className={className}>
        <tbody>
          {_.map(items, (v, k) => (
            <tr key={k}>
              <td>
                <Text capitalize bold>
                  {k}:
                </Text>
              </td>
              <td>{v}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }
)`
  tbody tr td:first-child {
    min-width: 200px;
  }
  tr {
    height: 16px;
  }
`;

function KustomizationDetail({ className, name }: Props) {
  const { data, isLoading, error } = useGetKustomization(name);

  const kustomization = data?.kustomization;

  return (
    <Page
      title={kustomization?.name}
      loading={isLoading}
      error={error}
      className={className}
    >
      <Info>
        <h3>{kustomization?.namespace}</h3>
        <InfoList
          items={{
            Source: (
              <Link
                to={formatURL(V2Routes.GitRepo, {
                  name: kustomization?.sourceRef.name,
                })}
              >
                GitRepository/{kustomization?.sourceRef.name}
              </Link>
            ),
            Status: (
              <Flex start>
                <KubeStatusIndicator conditions={kustomization?.conditions} />
                <div>
                  &nbsp; Applied revision {kustomization?.lastAppliedRevision}
                </div>
              </Flex>
            ),
            Cluster: "",
            Path: kustomization?.path,
            Interval: `${kustomization?.interval.hours}h ${kustomization?.interval.minutes}m`,
            "Last Updated At": kustomization?.lastHandledReconciledAt,
          }}
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
})`
  h3 {
    color: #737373;
    font-weight: 200;
    margin-top: 12px;
  }
`;
