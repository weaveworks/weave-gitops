import * as React from "react";
import styled from "styled-components";
import ActionBar from "../components/ActionBar";
import Button from "../components/Button";
import DataTable from "../components/DataTable";
import Flex from "../components/Flex";
import Icon, { IconType } from "../components/Icon";
import Link from "../components/Link";
import Page, { TitleBar } from "../components/Page";
import useApplications from "../hooks/applications";
import { Application } from "../lib/api/applications/applications.pb";
import { PageRoute } from "../lib/types";
import { formatURL } from "../lib/utils";
type Props = {
  className?: string;
};

function Applications({ className }: Props) {
  const [applications, setApplications] = React.useState<Application[]>([]);
  const { listApplications, loading } = useApplications();

  React.useEffect(() => {
    listApplications().then((res) => setApplications(res as Application[]));
  }, []);

  const AppButton = styled(Button)`
    &.MuiButton-root {
      border-color: #bdbdbd;
      font-weight: 600;
    }
  `;

  return (
    <Page loading={loading} className={className}>
      <Flex align between>
        <TitleBar>
          <h2>Installed GitOps Applications</h2>
        </TitleBar>
      </Flex>
      <ActionBar>
        <Link to={PageRoute.ApplicationAdd}>
          <AppButton variant="outlined" color="primary" type="button">
            Add a new app <Icon type={IconType.Add} size="base" />
          </AppButton>
        </Link>
        <AppButton variant="outlined" color="primary" type="button">
          Install a new Profile <Icon type={IconType.Add} size="base" />
        </AppButton>
        <AppButton variant="outlined" color="secondary" type="button">
          Create a PR to Delete App{" "}
          <Icon type={IconType.DeleteForever} size="base" />
        </AppButton>
      </ActionBar>
      <DataTable
        checks
        sortFields={["Name", "Type", "Namespace"]}
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
            value: "status",
          },
          {
            label: "Type",
            value: "type",
          },
          {
            label: "Version",
            value: "version",
          },
          {
            label: "Namespace",
            value: ({ namespace }: Application) => (
              <p>{namespace ? namespace : "default"}</p>
            ),
          },
          // Probably going to need this eventually, but we don't have a status
          // for an app from the backend yet. Keep the code around to avoid
          // having to figure this out again.
          // {
          //   label: "Status",
          //   value: () => (
          //     <Icon
          //       size="medium"
          //       color="success"
          //       type={IconType.CheckMark}
          //       text="Ready"
          //     />
          //   ),
          // },
        ]}
        rows={applications}
      />
    </Page>
  );
}

export default styled(Applications)``;
