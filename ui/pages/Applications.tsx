import * as React from "react";
import styled from "styled-components";
import DataTable from "../components/DataTable";
import Icon, { IconType } from "../components/Icon";
import Link from "../components/Link";
import Page from "../components/Page";
import Timestamp from "../components/Timestamp";
import { PageRoute } from "../lib/types";
import { formatURL } from "../lib/utils";

type Props = {
  className?: string;
};

const rows = [
  {
    name: "my-cool-app",
    status: "Ready",
    lastUpdate: "2006-01-02T15:04:05-0700",
  },
  { name: "podinfo", status: "Ready", lastUpdate: "2006-01-02T15:04:05-0700" },
  { name: "nginx", status: "Ready", lastUpdate: "2006-01-02T15:04:05-0700" },
];

function Applications({ className }: Props) {
  return (
    <Page title="Applications" className={className}>
      <DataTable
        sortFields={["name"]}
        fields={[
          {
            label: "Name",
            value: ({ name }) => (
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
          {
            label: "Last Updated",
            value: (v) => <Timestamp time={v.lastUpdate} />,
          },
        ]}
        rows={rows}
      />
    </Page>
  );
}

export default styled(Applications)``;
