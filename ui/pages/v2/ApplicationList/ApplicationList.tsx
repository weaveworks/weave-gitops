import * as React from "react";
import styled from "styled-components";
import DataTable, { SortType } from "../../../components/DataTable";
import KubeStatusIndicator from "../../../components/KubeStatusIndicator";
import Link from "../../../components/Link";
import Page from "../../../components/Page";
import { Automation, useListAutomations } from "../../../hooks/kustomizations";
import { AutomationType, V2Routes } from "../../../lib/types";
import { computeReady, formatURL } from "../../../lib/utils";

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
            sortType: SortType.string,
            sortValue: ({ name }) => name,
          },
          {
            label: "Type",
            value: "type",
            sortType: SortType.string,
            sortValue: ({ type }) => type,
          },
          {
            label: "Namespace",
            value: "namespace",
            sortType: SortType.string,
            sortValue: ({ namespace }) => namespace,
          },
          {
            label: "Cluster",
            value: "cluster",
            sortType: SortType.string,
            sortValue: ({ cluster }) => cluster,
          },
          {
            label: "Source",
            value: ({ sourceRef: { name, kind } }: Automation) => (
              <Link key={name} to={formatURL(V2Routes.Source, { name })}>
                {kind}/{name}
              </Link>
            ),
            sortType: SortType.string,
            sortValue: ({ sourceRef }) => sourceRef.name,
          },
          {
            label: "Status",
            value: ({ conditions }: Automation) => (
              <KubeStatusIndicator conditions={conditions} />
            ),
            sortType: SortType.bool,
            sortValue: ({ conditions }) => {
              const ready = computeReady(conditions) === "True";
              return ready ? true : false;
            },
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
            sortType: SortType.date,
            sortValue: ({ lastHandledReconciledAt }) => lastHandledReconciledAt,
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
