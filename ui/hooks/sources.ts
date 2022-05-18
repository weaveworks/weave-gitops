import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListBucketsResponse,
  ListGitRepositoriesResponse,
  ListHelmChartsResponse,
  ListHelmRepositoriesResponse,
} from "../lib/api/core/core.pb";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { NoNamespace, RequestError, Source } from "../lib/types";

export function useListSources(
  appName?: string,
  namespace: string = NoNamespace
) {
  const { api } = useContext(CoreClientContext);

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
            type: FluxObjectKind.KindGitRepository,
          })),
          ..._.map(hrs, (c) => ({
            ...c,
            type: FluxObjectKind.KindHelmRepository,
          })),
          ..._.map(buckets, (b) => ({
            ...b,
            type: FluxObjectKind.KindBucket,
          })),
          ..._.map(charts, (ch) => ({
            ...ch,
            type: FluxObjectKind.KindHelmChart,
          })),
        ];
      });
    },
    { retry: false, refetchInterval: 5000 }
  );
}
