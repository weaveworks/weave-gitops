import * as React from "react";
import styled from "styled-components";
import Button from "../../../components/Button";
import DataTable from "../../../components/DataTable";
import Link from "../../../components/Link";
import Page from "../../../components/Page";
import { useListApplications } from "../../../hooks/apps";
import { App } from "../../../lib/api/app/apps.pb";
import { V2Routes } from "../../../lib/types";
import { formatURL } from "../../../lib/utils";
import AppOnboarding from "./AppOnboarding";

type Props = {
  className?: string;
};

function ApplicationList({ className }: Props) {
  const { data = { apps: [] }, isLoading, error } = useListApplications();

  return (
    <Page
      title="Applications"
      error={error?.code === 13 ? null : error}
      loading={!error && isLoading}
      className={className}
      actions={
        <Link to={formatURL(V2Routes.NewApp)}>
          <Button>Add Application</Button>
        </Link>
      }
    >
      {error || data?.apps.length === 0 ? (
        <AppOnboarding />
      ) : (
        <DataTable
          sortFields={["name"]}
          fields={[
            {
              label: "Name",
              value: ({ name }: App) => (
                <Link key={name} to={formatURL(V2Routes.Application, { name })}>
                  {name}
                </Link>
              ),
            },
          ]}
          rows={data?.apps}
        />
      )}
    </Page>
  );
}

export default styled(ApplicationList).attrs({
  className: ApplicationList.name,
})``;
