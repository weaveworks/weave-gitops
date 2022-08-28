import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListBucketsResponse,
  ListGitRepositoriesResponse,
  ListHelmChartsResponse,
  ListHelmRepositoriesResponse,
  ListOCIRepositoriesResponse,
} from "../lib/api/core/core.pb";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import {
  MultiRequestError,
  NoNamespace,
  ReactQueryOptions,
  RequestError,
  Source,
} from "../lib/types";

type Res = { result: Source[]; errors: MultiRequestError[] };

export function useListSources(
  appName?: string,
  namespace: string = NoNamespace,
  opts: ReactQueryOptions<Res, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<Res, RequestError>(
    "sources",
    () => {
      const p = [
        api.ListGitRepositories({ namespace }),
        api.ListHelmRepositories({ namespace }),
        api.ListBuckets({ namespace }),
        api.ListHelmCharts({ namespace }),
        api.ListOCIRepositories({ namespace }),
      ];
      return Promise.all<any>(p).then((result) => {
        const [
          repoRes,
          helmRepositories,
          bucketsRes,
          chartRes,
          ociRepositories,
        ] = result;
        const repos = (repoRes as ListGitRepositoriesResponse).gitRepositories;
        const hrs = (helmRepositories as ListHelmRepositoriesResponse)
          .helmRepositories;
        const buckets = (bucketsRes as ListBucketsResponse).buckets;
        const charts = (chartRes as ListHelmChartsResponse).helmCharts;
        const ocis = (ociRepositories as ListOCIRepositoriesResponse)
          .ociRepositories;
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
            ..._.map(ocis, (c) => ({
              ...c,
              kind: FluxObjectKind.KindOCIRepository,
            })),
          ],
          errors: [
            ..._.map(repoRes.errors, (e) => ({
              ...e,
              kind: FluxObjectKind.KindGitRepository,
            })),
            ..._.map(helmRepositories.errors, (e) => ({
              ...e,
              kind: FluxObjectKind.KindHelmRepository,
            })),
            ..._.map(bucketsRes.errors, (e) => ({
              ...e,
              kind: FluxObjectKind.KindBucket,
            })),
            ..._.map(chartRes.errors, (e) => ({
              ...e,
              kind: FluxObjectKind.KindHelmChart,
            })),
            ..._.map(ociRepositories.errors, (e) => ({
              ...e,
              kind: FluxObjectKind.KindOCIRepository,
            })),
          ],
        };
      });
    },
    opts
  );
}
