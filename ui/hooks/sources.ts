import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  ListGitRepositoriesResponse,
  ListHelmChartsResponse,
  ListHelmRepositoriesResponse,
} from "../lib/api/core/core.pb";
import { SourceRefSourceKind } from "../lib/api/core/types.pb";
import { RequestError, Source, WeGONamespace } from "../lib/types";

export function useListSources(
  appName?: string,
  namespace: string = WeGONamespace
) {
  const { api } = useContext(AppContext);

  return useQuery<Source[], RequestError>(
    "sources",
    () => {
      const p = [
        api.ListGitRepositories({ namespace }),
        api.ListHelmRepositories({ namespace }),
        api.ListHelmCharts({ namespace }),
      ];
      return Promise.all(p).then((result) => {
        const [repoRes, helmReleases, chartRes] = result;
        const repos = (repoRes as ListGitRepositoriesResponse).gitRepositories;
        const hrs = (helmReleases as ListHelmRepositoriesResponse)
          .helmRepositories;
        const charts = (chartRes as ListHelmChartsResponse).helmCharts;

        return [
          ..._.map(repos, (r) => ({
            ...r,
            type: SourceRefSourceKind.GitRepository,
          })),
          ..._.map(hrs, (c) => ({
            ...c,
            type: SourceRefSourceKind.HelmRepository,
          })),
          ..._.map(charts, (c) => ({
            ...c,
            type: SourceRefSourceKind.HelmChart,
          })),
        ];
      });
    },
    { retry: false }
  );
}
