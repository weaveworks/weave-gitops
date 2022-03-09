import _ from "lodash";
import { useContext, useEffect } from "react";
import { AppContext } from "../contexts/AppContext";
import {
  ParseRepoURLResponse,
} from "../lib/api/applications/applications.pb";
import { useRequestState } from "./common";

export function useParseRepoURL(url: string) {
  const { applicationsClient } = useContext(AppContext);
  const [res, loading, error, req] = useRequestState<ParseRepoURLResponse>();

  useEffect(() => {
    req(applicationsClient.ParseRepoURL({ url }));
  }, [url]);

  return [res, loading, error];
}
