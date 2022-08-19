import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import {
  Bucket,
  FluxObjectKind,
  GitRepository,
  HelmRepository,
  OCIRepository,
} from "../lib/api/core/types.pb";
import { formatURL, objectTypeToRoute } from "../lib/nav";
import { showInterval } from "../lib/time";
import { Source } from "../lib/types";
import {
  convertGitURLToGitProvider,
  removeKind,
  statusSortHelper,
} from "../lib/utils";
import { Field } from "./DataTable";
import FilterableTable, {
  filterByStatusCallback,
  filterConfig,
} from "./FilterableTable";
import KubeStatusIndicator, { computeMessage } from "./KubeStatusIndicator";
import Link from "./Link";
import Timestamp from "./Timestamp";

type Props = {
  className?: string;
  sources?: Source[];
  appName?: string;
};

function SourcesTable({ className, sources }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  const [filterDialogOpen, setFilterDialog] = React.useState(false);
  sources = sources?.map((s) => {
    return { ...s, type: removeKind(s.kind) };
  });

  let initialFilterState = {
    ...filterConfig(sources, "type"),
    ...filterConfig(sources, "namespace"),
    ...filterConfig(sources, "status", filterByStatusCallback),
  };

  if (flags.WEAVE_GITOPS_FEATURE_TENANCY === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(sources, "tenant"),
    };
  }

  if (flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(sources, "clusterName"),
    };
  }

  const fields: Field[] = [
    {
      label: "Name",
      value: (s: Source) => (
        <Link
          to={formatURL(objectTypeToRoute(s.kind), {
            name: s?.name,
            namespace: s?.namespace,
            clusterName: s?.clusterName,
          })}
        >
          {s?.name}
        </Link>
      ),
      sortValue: (s: Source) => s.name || "",
      textSearchable: true,
      maxWidth: 600,
    },
    { label: "Type", value: "type" },
    { label: "Namespace", value: "namespace" },
    ...(flags.WEAVE_GITOPS_FEATURE_TENANCY === "true"
      ? [{ label: "Tenant", value: "tenant" }]
      : []),
    ...(flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [{ label: "Cluster", value: (s: Source) => s.clusterName }]
      : []),
    {
      label: "Status",
      value: (s: Source) => (
        <KubeStatusIndicator
          short
          conditions={s.conditions}
          suspended={s.suspended}
        />
      ),
      sortValue: statusSortHelper,
      defaultSort: true,
    },
    {
      label: "Message",
      value: (s) => computeMessage(s.conditions),
      maxWidth: 600,
    },
    {
      label: "URL",
      value: (s: Source) => {
        let text;
        let url;
        let link = false;
        switch (s.kind) {
          case FluxObjectKind.KindGitRepository:
            text = (s as GitRepository).url;
            url = convertGitURLToGitProvider((s as GitRepository).url);
            link = true;
            break;
          case FluxObjectKind.KindBucket:
            text = (s as Bucket).endpoint;
            break;
          case FluxObjectKind.KindOCIRepository:
            text = (s as OCIRepository).url;
            break;
          case FluxObjectKind.KindHelmRepository:
            text = (s as HelmRepository).url;
            url = text;
            link = true;
            break;
          default:
            text = "-";
            break;
        }
        return link ? (
          <Link newTab href={url}>
            {text}
          </Link>
        ) : (
          text
        );
      },
      maxWidth: 600,
    },
    {
      label: "Reference",
      value: (s: Source) => {
        if (s.kind === FluxObjectKind.KindGitRepository) {
          const repo = s as GitRepository;
          const ref =
            repo?.reference?.branch ||
            repo?.reference?.commit ||
            repo?.reference?.tag ||
            repo?.reference?.semver;
          return ref;
        }
        return "-";
      },
    },
    {
      label: "Interval",
      value: (s: Source) => showInterval(s.interval),
    },
    {
      label: "Last Updated",
      value: (s: Source) =>
        s.lastUpdatedAt ? <Timestamp time={s.lastUpdatedAt} /> : "-",
      sortValue: (s: Source) => s.lastUpdatedAt || "",
    },
  ];

  return (
    <FilterableTable
      className={className}
      filters={initialFilterState}
      rows={sources}
      dialogOpen={filterDialogOpen}
      onDialogClose={() => setFilterDialog(false)}
      fields={fields}
    />
  );
}

export default styled(SourcesTable).attrs({ className: SourcesTable.name })`
  td:nth-child(5) {
    white-space: pre-wrap;
    overflow-wrap: break-word;
    word-wrap: break-word;
  }
`;
