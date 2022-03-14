import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  ListBucketsResponse,
  ListGitRepositoriesResponse,
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
        api.ListBuckets({ namespace }),
      ];
      return Promise.all(p).then((result) => {
        const [repoRes, helmReleases, bucketsRes] = result;
        const repos = (repoRes as ListGitRepositoriesResponse).gitRepositories;
        const hrs = (helmReleases as ListHelmRepositoriesResponse)
          .helmRepositories;
        const buckets = (bucketsRes as ListBucketsResponse).buckets;

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
        ];
      });
    },
    { retry: false }
  );
}
