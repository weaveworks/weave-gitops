import * as React from "react";
import styled from "styled-components";
import DataTable from "../components/DataTable";
import Icon, { IconType } from "../components/Icon";
import Link from "../components/Link";
import Page from "../components/Page";
import useApplications from "../hooks/applications";
import { Application } from "../lib/api/applications/applications.pb";
import { PageRoute } from "../lib/types";
import { formatURL } from "../lib/utils";

type Props = {
  className?: string;
};

function Applications({ className }: Props) {
  const { applications } = useApplications();

  return (
    <Page title="Applications" className={className}>
      <DataTable
        sortFields={["name"]}
        fields={[
          {
            label: "Name",
            value: ({ name }: Application) => (
              <Link to={formatURL(PageRoute.ApplicationDetail, { name })}>
                {name}
              </Link>
            ),
          },
          {
            label: "Status",
            value: () => (
              <Icon
                size="medium"
                color="success"
                type={IconType.CheckMark}
                text="Ready"
              />
            ),
          },
        ]}
        rows={applications}
      />
    </Page>
  );
}

export default styled(Applications)``;
