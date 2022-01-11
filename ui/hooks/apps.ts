import { useContext } from "react";
import { useMutation, useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  AddAppRequest,
  AddAppResponse,
  GetAppResponse,
  ListAppResponse,
  RemoveAppRequest,
  RemoveAppResponse,
} from "../lib/api/app/apps.pb";
import { RequestError } from "../lib/types";
import { useUserConfigRepoName } from "./common";

export function useListApplications() {
  const { apps, doAsyncError } = useContext(AppContext);
  const repoName = useUserConfigRepoName();

  return useQuery<ListAppResponse, RequestError>(
    "apps",
    () =>
      apps.ListApps({ repoName }).catch((e) => {
        doAsyncError(e.message, e.detail);
        throw e;
      }),
    {
      retry: false,
    }
  );
}

export function useGetApplication(appName: string) {
  const { apps } = useContext(AppContext);
  const repoName = useUserConfigRepoName();

  return useQuery<GetAppResponse, RequestError>(
    ["apps", appName],
    () => apps.GetApp({ repoName, appName }),
    { retry: false }
  );
}

export function useCreateApp() {
  const { apps } = useContext(AppContext);
  const repoName = useUserConfigRepoName();

  const mutation = useMutation<AddAppResponse, RequestError, AddAppRequest>(
    (body: AddAppRequest) => apps.AddApp({ ...body, repoName })
  );

  return mutation;
}

export function useRemoveApp() {
  const { apps } = useContext(AppContext);
  const repoName = useUserConfigRepoName();

  return useMutation<RemoveAppResponse, RequestError, RemoveAppRequest>(
    (body: RemoveAppRequest) => apps.RemoveApp({ ...body, repoName })
  );
}
