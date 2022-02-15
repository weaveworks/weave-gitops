import * as React from "react";
import styled from "styled-components";
import DataTable from "../../../components/DataTable";
import KubeStatusIndicator from "../../../components/KubeStatusIndicator";
import Link from "../../../components/Link";
import Page from "../../../components/Page";
import { Automation, useListAutomations } from "../../../hooks/kustomizations";
import { AutomationType, V2Routes } from "../../../lib/types";
import { formatURL } from "../../../lib/utils";

type Props = {
  className?: string;
};

function ApplicationList({ className }: Props) {
  const { data, isLoading, error } = useListAutomations();

  return (
    <Page
      title="Releases"
      error={error}
      loading={!error && isLoading}
      className={className}
    >
      <DataTable
        sortFields={[
          "name",
          "namespace",
          "type",
          "cluster",
          "status",
          "last synced at",
        ]}
        fields={[
          {
            label: "Name",
            value: ({ name, type }: Automation) => (
              <Link
                key={name}
                to={formatURL(
                  type === AutomationType.Kustomization
                    ? V2Routes.Kustomization
                    : V2Routes.HelmRelease,
                  { name }
                )}
              >
                {name}
              </Link>
            ),
          },
          {
            label: "Type",
            value: "type",
          },
          {
            label: "Namespace",
            value: "namespace",
          },
          {
            label: "Cluster",
            value: "cluster",
          },
          {
            label: "Source",
            value: ({ sourceRef: { name, kind } }: Automation) => (
              <Link key={name} to={formatURL(V2Routes.Source, { name })}>
                {kind}/{name}
              </Link>
            ),
          },
          {
            label: "Status",
            value: ({ conditions }: Automation) => (
              <KubeStatusIndicator conditions={conditions} />
            ),
          },
          {
            label: "Release",
            value: (v: Automation) => {
              if (!v.lastAppliedRevision) {
                return null;
              }
              const [branch, hash] = v.lastAppliedRevision.split("/");

              return `${branch}/${hash.substring(0, 7)}`;
            },
          },
          {
            label: "Last Synced At",
            value: "lastHandledReconciledAt",
          },
        ]}
        rows={data}
      />
    </Page>
  );
}

export default styled(ApplicationList).attrs({
  className: ApplicationList.name,
})``;
