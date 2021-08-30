import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import ConditionsTable from "../components/ConditionsTable";
import DataTable from "../components/DataTable";
import KeyValueTable from "../components/KeyValueTable";
import Page from "../components/Page";
import { AppContext } from "../contexts/AppContext";
import useApplications from "../hooks/applications";
import {
  Application,
  UnstructuredObject,
} from "../lib/api/applications/applications.pb";
import { getChildren } from "../lib/graph";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
  name: string;
};

function ApplicationDetail({ className, name }: Props) {
  const [app, setApp] = React.useState<Application>({});
  const { applicationsClient } = React.useContext(AppContext);
  const [reconciledObjects, setReconciledObjects] = React.useState<
    UnstructuredObject[]
  >([]);

  const { getApplication, loading } = useApplications();

  React.useEffect(() => {
    getApplication(name).then((app) => setApp(app || {}));
  }, []);

  React.useEffect(() => {
    if (!app) {
      return;
    }

    const uniqKinds = _.uniqBy(app.reconciledObjectKinds, "kind");
    if (uniqKinds.length === 0) {
      return;
    }

    getChildren(applicationsClient, app, uniqKinds).then((objs) =>
      setReconciledObjects(objs)
    );
  }, [app]);

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
      <h3>Reconciled Objects</h3>
      <DataTable
        sortFields={["name"]}
        fields={[
          { label: "Name", value: "name" },
          { label: "Kind", value: (v) => v.groupVersionKind.kind },
          { label: "Status", value: "status" },
        ]}
        rows={reconciledObjects}
      />
    </Page>
  );
}

export default styled(ApplicationDetail)``;
