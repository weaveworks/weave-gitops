import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListBucketsResponse,
  ListError,
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

  return useQuery<{ result: Source[]; errors: ListError[] }, RequestError>(
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
        const ErrorList: ListError[] = [
          ...repoRes.errors,
          ...helmReleases.errors,
          ...bucketsRes.errors,
          ...chartRes.errors,
        ];
        return {
          result: [
            ..._.map(repos, (r) => ({
              ...r,
              kind: FluxObjectKind.KindGitRepository,
            })),
            ..._.map(hrs, (c) => ({
              ...c,
              kind: FluxObjectKind.KindHelmRepository,
            })),
            ..._.map(buckets, (b) => ({
              ...b,
              kind: FluxObjectKind.KindBucket,
            })),
            ..._.map(charts, (ch) => ({
              ...ch,
              kind: FluxObjectKind.KindHelmChart,
            })),
          ],
          errors: _.uniqBy(ErrorList, "clusterName"),
        };
      });
    },
    { retry: false, refetchInterval: 5000 }
  );
}
