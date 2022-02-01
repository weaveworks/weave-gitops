import { CircularProgress } from "@material-ui/core";
import { DateTime } from "luxon";
import * as React from "react";
import styled from "styled-components";
import { useListCommits } from "../hooks/applications";
import {
  Application,
  Commit,
  GitProvider,
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
  authSuccess: boolean;
  onAuthClick?: () => void;
  provider: string;
};

const timestamp = (isoStr: string) => {
  if (process.env.NODE_ENV === "test") {
    return "test timestamp";
  }

  return DateTime.fromISO(isoStr).toRelative();
};

function CommitsTable({
  className,
  app,
  authSuccess,
  onAuthClick,
  provider,
}: Props) {
  const [commits, loading, error, req] = useListCommits();

  React.useEffect(() => {
    if (!app || !app.name) {
      return;
    }

    req(GitProvider[provider], {
      name: app.name,
      namespace: app.namespace,
      pageSize: 10,
    });
  }, [app, authSuccess]);

  if (error) {
    return error.code === GrpcErrorCodes.Unauthenticated ? (
      <AuthAlert
        provider={GitProvider[provider]}
        title="Error fetching commits"
        onClick={onAuthClick}
      />
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
        fields={[
          {
            label: "SHA",
            value: (row: Commit) => (
              <Link newTab href={row.url}>
                {row.hash}
              </Link>
            ),
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
