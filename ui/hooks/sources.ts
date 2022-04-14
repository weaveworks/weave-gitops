import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  ListBucketsResponse,
  ListGitRepositoriesResponse,
  ListHelmChartsResponse,
  ListHelmRepositoriesResponse,
} from "../lib/api/core/core.pb";
import { SourceRefSourceKind } from "../lib/api/core/types.pb";
import { NoNamespace, RequestError, Source } from "../lib/types";

export function useListSources(
  appName?: string,
  namespace: string = NoNamespace
) {
  const { api } = useContext(AppContext);

  return useQuery<Source[], RequestError>(
    "sources",
    () => {
      const p = [
        api.ListGitRepositories({ namespace }),
        api.ListHelmRepositories({ namespace }),
        api.ListBuckets({ namespace }),
        api.ListHelmCharts({ namespace }),
      ];
      return Promise.all<any>(p).then((result) => {
        const [repoRes, helmReleases, bucketsRes, chartRes] = result;
        const repos = (repoRes as ListGitRepositoriesResponse).gitRepositories;
        const hrs = (helmReleases as ListHelmRepositoriesResponse)
          .helmRepositories;
        const buckets = (bucketsRes as ListBucketsResponse).buckets;
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
          ..._.map(buckets, (b) => ({
            ...b,
            type: SourceRefSourceKind.Bucket,
          })),
          ..._.map(charts, (ch) => ({
            ...ch,
            type: SourceRefSourceKind.HelmChart,
          })),
        ];
      });
    },
    { retry: false, refetchInterval: 5000 }
  );
}
