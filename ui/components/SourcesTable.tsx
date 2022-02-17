import * as React from "react";
import styled from "styled-components";
import { GitRepository, HelmChart, SourceType } from "../lib/api/app/source.pb";
import { Source, V2Routes } from "../lib/types";
import { convertGitURLToGitProvider, formatURL } from "../lib/utils";
import DataTable from "./DataTable";
import KubeStatusIndicator from "./KubeStatusIndicator";
import Link from "./Link";

type Props = {
  className?: string;
  sources: Source[];
  appName?: string;
};

function SourcesTable({ className, sources }: Props) {
  return (
    <div className={className}>
      <DataTable
        sortFields={["name"]}
        rows={sources}
        fields={[
          {
            label: "Name",
            value: (s: Source) => (
              <Link
                to={formatURL(V2Routes.Source, {
                  name: s?.name,
                  namespace: s?.namespace,
                })}
              >
                {s?.name}
              </Link>
            ),
          },
          { label: "Type", value: "type" },

          {
            label: "Status",
            value: (s: Source) => (
              <KubeStatusIndicator conditions={s.conditions} />
            ),
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

              if (s.type === SourceType.GitRepository) {
                text = (s as GitRepository).url;
                url = convertGitURLToGitProvider((s as GitRepository).url);
              } else {
                text = `https://${(s as HelmChart).sourceRef?.name}`;
                url = (s as HelmChart).chart;
              }

              return (
                <Link newTab href={url}>
                  {text}
                </Link>
              );
            },
          },
          {
            label: "Reference",
            value: (s: Source) => {
              const isGit = s.type === SourceType.GitRepository;
              const repo = s as GitRepository;
              const ref =
                repo.reference.branch ||
                repo.reference.commit ||
                repo.reference.tag ||
                repo.reference.semver;

              return isGit ? ref : "";
            },
          },
          {
            label: "Interval",
            value: (s: Source) =>
              `${s.interval.hours}h${s.interval.minutes}m${s.interval.seconds}s`,
          },
        ]}
      />
    </div>
  );
}

export default styled(SourcesTable).attrs({ className: SourcesTable.name })``;
