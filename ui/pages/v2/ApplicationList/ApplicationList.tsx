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
  const placeholder: App[] = [
    {
      name: "my-app",
      displayName: "My App",
      description: "This application runs our Valorant match making back-end",
      kustomizations: [
        // { name: "my-kustomization" },
      ],
      helmReleases: [],
      sources: [],
    },
  ];
  const { data = {}, isLoading, error } = useListApplications();

  data.apps = placeholder as any;

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
              value: ({ id, name }: App) => (
                <Link key={id} to={formatURL(V2Routes.Application, { name })}>
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
