import _ from "lodash";
import { useContext } from "react";
import { useMutation, useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  AddAppRequest,
  AddAppResponse,
  CreateDeployKeyRequest,
  CreateDeployKeyResponse,
  GetAppResponse,
  ListAppResponse,
  RemoveAppRequest,
  RemoveAppResponse,
} from "../lib/api/app/apps.pb";
import { RequestError, WeGONamespace } from "../lib/types";
import { makeHeaders } from "./common";

export function useListApplications() {
  const { apps, doAsyncError } = useContext(AppContext);

  return useQuery<ListAppResponse, RequestError>(
    "apps",
    () =>
      apps.ListApps({ namespace: WeGONamespace }).catch((e) => {
        doAsyncError(e.message, e.detail);
        throw e;
      }),
    {
      retry: false,
    }
  );
}

export function useGetApplication(
  appName: string,
  namespace: string = WeGONamespace
) {
  const { apps } = useContext(AppContext);

  return useQuery<GetAppResponse, RequestError>(
    ["apps", appName],
    () => apps.GetApp({ appName, namespace }),
    { retry: false }
  );
}

export function useCreateApp() {
  const { apps } = useContext(AppContext);

  const mutation = useMutation<AddAppResponse, RequestError, AddAppRequest>(
    (body: AddAppRequest) => apps.AddApp({ ...body })
  );

  return mutation;
}

export function useRemoveApp() {
  const { apps } = useContext(AppContext);

  return useMutation<RemoveAppResponse, RequestError, RemoveAppRequest>(
    (body: RemoveAppRequest) => apps.RemoveApp({ ...body })
  );
}

export function useCreateDeployKey() {
  const { apps } = useContext(AppContext);

  const { getProviderToken } = useContext(AppContext);

  return useMutation<
    CreateDeployKeyResponse,
    RequestError,
    CreateDeployKeyRequest
  >((body: CreateDeployKeyRequest) => {
    const headers = makeHeaders(_.bind(getProviderToken, this, body.provider));
    return apps.CreateDeployKey(body, { headers });
  });
}
