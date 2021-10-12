import { CircularProgress } from "@material-ui/core";
import { DateTime } from "luxon";
import * as React from "react";
import { useContext } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { isUnauthenticated } from "../hooks/auth";
import { useRequestState } from "../hooks/common";
import {
  Application,
  Commit,
  ListCommitsResponse,
} from "../lib/api/applications/applications.pb";
import { GrpcErrorCodes } from "../lib/types";
import Alert from "./Alert";
import AuthAlert from "./AuthAlert";
import DataTable from "./DataTable";
import Flex from "./Flex";
import Link from "./Link";

type Props = {
  className?: string;
  app: Application;
  onAuthClick?: () => void;
};

const timestamp = (isoStr: string) => {
  if (process.env.NODE_ENV === "test") {
    return "test timestamp";
  }

  return DateTime.fromISO(isoStr).toRelative();
};

function CommitsTable({ className, app, onAuthClick }: Props) {
  const { applicationsClient } = useContext(AppContext);
  const [commits, loading, error, req] = useRequestState<ListCommitsResponse>();

  React.useEffect(() => {
    if (!app || !app.name) {
      return;
    }

    req(
      applicationsClient.ListCommits({
        name: app.name,
        namespace: app.namespace,
        pageSize: 10,
      })
    );
  }, [app]);

  if (error) {
    return isUnauthenticated(error.code) ? (
      <AuthAlert title="Error fetching commits" onClick={onAuthClick} />
    ) : (
      <Alert
        className={className}
        severity="error"
        title="Error fetching commits"
        message={error.message}
      />
    );
  }

  if ((!commits && !error) || loading) {
    return (
      <Flex wide center>
        <CircularProgress />
      </Flex>
    );
  }

  return (
    <div className={className}>
      <DataTable
        sortFields={["date"]}
        reverseSort
        fields={[
          {
            label: "SHA",
            value: (row: Commit) => <Link href={row.url}>{row.hash}</Link>,
          },
          {
            label: "Date",
            value: (row: Commit) => timestamp(row.date),
          },
          { label: "Message", value: "message" },
          { label: "Author", value: "author" },
        ]}
        rows={commits.commits}
      />
    </div>
  );
}

export default styled(CommitsTable)``;
