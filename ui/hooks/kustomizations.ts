import { useContext } from "react";
import { useMutation } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  AddKustomizationRequest,
  AddKustomizationResponse,
} from "../lib/api/app/kustomize.pb";
import { RequestError } from "../lib/types";
import { useUserConfigRepoName } from "./common";

export function useCreateKustomization() {
  const { kustomizations } = useContext(AppContext);

  const repoName = useUserConfigRepoName();

  return useMutation<
    AddKustomizationResponse,
    RequestError,
    AddKustomizationRequest
  >((body: AddKustomizationRequest) =>
    kustomizations.Add({ ...body, repoName })
  );
}
