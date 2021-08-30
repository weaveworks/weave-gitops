import * as React from "react";
import styled from "styled-components";
import ConditionsTable from "../components/ConditionsTable";
import KeyValueTable from "../components/KeyValueTable";
import Page from "../components/Page";
import useApplications from "../hooks/applications";
import { Application } from "../lib/api/applications/applications.pb";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
  name: string;
};

function ApplicationDetail({ className, name }: Props) {
  const [app, setApp] = React.useState<Application>({});

  const { getApplication, loading } = useApplications();

  React.useEffect(() => {
    getApplication(name).then((app) => setApp(app || {}));
  }, []);

  return (
    <Page
      loading={loading}
      breadcrumbs={[{ page: PageRoute.Applications }]}
      title={name}
      className={className}
    >
      <KeyValueTable
        columns={4}
        pairs={[
          { key: "Name", value: app.name },
          { key: "URL", value: app.url },
          { key: "Path", value: app.path },
        ]}
      />
      <h3>Source Conditions</h3>
      <ConditionsTable conditions={app.sourceConditions} />
      <h3>Automation Conditions</h3>
      <ConditionsTable conditions={app.deploymentConditions} />
    </Page>
  );
}

export default styled(ApplicationDetail)``;
