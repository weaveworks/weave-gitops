import * as React from "react";
import styled from "styled-components";
import {
  Bucket,
  GitRepository,
  HelmChart,
  SourceRefSourceKind,
} from "../lib/api/core/types.pb";
import { formatURL, sourceTypeToRoute } from "../lib/nav";
import { Source } from "../lib/types";
import { convertGitURLToGitProvider } from "../lib/utils";
import { showInterval } from "../lib/time";
import DataTable, { SortType } from "./DataTable";
import Flex from "./Flex";
import KubeStatusIndicator from "./KubeStatusIndicator";
import Link from "./Link";
import Text from "./Text";

type Props = {
  className?: string;
  sources: Source[];
  appName?: string;
};

const statusWidth = 480;

function SourcesTable({ className, sources }: Props) {
  return (
    <div className={className}>
      <DataTable
        rows={sources}
        fields={[
          {
            label: "Name",
            value: (s: Source) => (
              <Link
                to={formatURL(sourceTypeToRoute(s.type), {
                  name: s?.name,
                  namespace: s?.namespace,
                })}
              >
                {s?.name}
              </Link>
            ),
            sortType: SortType.string,
            sortValue: (s: Source) => s.name || "",
            width: 100,
          },
          { label: "Type", value: "type" },

          {
            label: "Status",
            value: (s: Source) => (
              <KubeStatusIndicator conditions={s.conditions} />
            ),
            width: statusWidth,
          },
          {
            label: "Cluster",
            value: "cluster",
          },
          {
            label: "URL",
            value: (s: Source) => {
              let text;
              let url;
              let link = false;

              if (s.type === SourceRefSourceKind.GitRepository) {
                text = (s as GitRepository).url;
                url = convertGitURLToGitProvider((s as GitRepository).url);
                link = true;
              } else if (s.type === SourceRefSourceKind.Bucket) {
                text = (s as Bucket).endpoint;
              } else if (s.type === SourceRefSourceKind.HelmChart) {
                text = `https://${(s as HelmChart).sourceRef?.name}`;
                url = (s as HelmChart).chart;
                link = true;
              }

              return link ? (
                <Link newTab href={url}>
                  {text}
                </Link>
              ) : (
                text
              );
            },
          },
          {
            label: "Reference",
            value: (s: Source) => {
              const isGit = s.type === SourceRefSourceKind.GitRepository;
              const repo = s as GitRepository;
              const ref =
                repo?.reference?.branch ||
                repo?.reference?.commit ||
                repo?.reference?.tag ||
                repo?.reference?.semver;

              return isGit ? ref : "-";
            },
          },
          {
            label: "Interval",
            value: (s: Source) => showInterval(s.interval)
          },
        ]}
      />
    </div>
  );
}

export default styled(SourcesTable).attrs({ className: SourcesTable.name })`
  /* Setting this here to get the ellipsis to work */
  /* Because this is a div within a td, overflow doesn't apply to the td */
  ${KubeStatusIndicator} ${Flex} ${Text} {
    max-width: ${statusWidth}px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
`;
