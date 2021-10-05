import * as React from "react";
import styled from "styled-components";
import Button from "../components/Button";
import DataTable from "../components/DataTable";
import Flex from "../components/Flex";
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

  return (
    <Page loading={loading} className={className}>
      <Flex align wide between>
        <TitleBar>
          <h2>Applications</h2>
        </TitleBar>
        <Link to={PageRoute.ApplicationAdd}>
          <Button variant="contained" color="primary" type="button">
            Add Application
          </Button>
        </Link>
      </Flex>
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
