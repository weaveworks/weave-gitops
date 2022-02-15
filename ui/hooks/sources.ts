import _ from "lodash";
import { useContext } from "react";
import { useMutation, useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  AddGitRepositoryReq,
  AddGitRepositoryRes,
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

  return useQuery<Source[], RequestError>(
    "sources",
    () => {
      const p = [
        apps.ListGitRepositories({ appName, namespace }),
        apps.ListHelmCharts({ appName, namespace }),
      ];
      return Promise.all(p).then((result) => {
        const [repoRes, chartRes] = result;
        const repos = (repoRes as ListGitRepositoryRes).gitRepositories;
        const charts = (chartRes as ListHelmChartRes).helmCharts;

        return [
          ..._.map(repos, (r) => ({
            ...r,
            type: SourceType.GitRepository,
          })),
          ..._.map(charts, (c) => ({
            ...c,
            type: SourceType.HelmChart,
          })),
        ];
      });
    },
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

export function useCreateRepo() {
  const { apps } = useContext(AppContext);

  return useMutation<AddGitRepositoryRes, RequestError, AddGitRepositoryReq>(
    (body: AddGitRepositoryReq) => apps.AddGitRepository(body)
  );
}
