import * as React from "react";
import styled from "styled-components";
import Button from "../components/Button";
import DataTable from "../components/DataTable";
import Flex from "../components/Flex";
import Icon, { IconType } from "../components/Icon";
import Link from "../components/Link";
import Page from "../components/Page";
import PollingIndicator from "../components/PollingIndicator";
import Spacer from "../components/Spacer";
import { AppContext } from "../contexts/AppContext";
import { Application } from "../lib/api/applications/applications.pb";
import { PageRoute } from "../lib/types";
import { formatURL } from "../lib/utils";

type Props = {
  className?: string;
};

function Applications({ className }: Props) {
  const [applications, setApplications] = React.useState<Application[]>([]);
  const { applicationsClient, doAsyncError } = React.useContext(AppContext);
  const [loading, setLoading] = React.useState(false);

  const getApps = () => {
    setLoading(true);
    applicationsClient
      .ListApplications({ namespace: "wego-system" })
      .then((res) => setApplications(res.applications))
      .catch((err) => doAsyncError(err.message, err.detail))
      .finally(() => setLoading(false));
  };

  React.useEffect(() => {
    getApps();
    const interval = setInterval(() => {
      getApps();
    }, 30000);
    return () => clearInterval(interval);
  }, []);

  return (
    <Page className={className}>
      <Flex between align>
        <Flex align>
          <h2>Applications</h2>
          <Spacer padding="small" />
          <PollingIndicator loading={loading} interval="30" />
        </Flex>
        <Link to={PageRoute.ApplicationAdd} className="title-bar-button">
          <Button
            startIcon={<Icon type={IconType.AddIcon} size="base" />}
            type="button"
          >
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
