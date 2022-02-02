import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  ListGitRepositoryRes,
  ListHelmChartRes,
  SourceType,
} from "../lib/api/app/source.pb";
import { RequestError, Source, WeGONamespace } from "../lib/types";

export function useListSources(
  appName?: string,
  namespace: string = WeGONamespace
) {
  const { apps } = useContext(AppContext);

  const p = [
    apps.ListGitRepositories({ appName, namespace }),
    apps.ListHelmCharts({ appName, namespace }),
  ];

  return useQuery<Source[], RequestError>(
    "sources",
    () =>
      Promise.all(p).then((result) => {
        const [repoRes, chartRes] = result;
        const repos = (repoRes as ListGitRepositoryRes).gitRepositories;
        const charts = (chartRes as ListHelmChartRes).helmCharts;

        return [
          ..._.map(repos, (r) => ({ name: r.name, type: SourceType.Git })),
          ..._.map(charts, (c) => ({ name: c.name, type: SourceType.Helm })),
        ];
      }),
    { retry: false }
  );
}

export function useListGitRepos(
  appName?: string,
  namespace: string = WeGONamespace
) {
  const { apps } = useContext(AppContext);

  return useQuery("gitrepos", () =>
    apps.ListGitRepositories({ appName, namespace })
  );
}
