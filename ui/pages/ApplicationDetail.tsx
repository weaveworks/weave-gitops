import * as React from "react";
import styled from "styled-components";
import KeyValueTable from "../components/KeyValueTable";
import Page from "../components/Page";
import useApplications from "../hooks/applications";
import useNavigation from "../hooks/navigation";
import { Application } from "../lib/api/applications/applications.pb";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
};

function ApplicationDetail({ className }: Props) {
  const [app, setApp] = React.useState<Application>({});
  const {
    query: { name },
  } = useNavigation<{ name: string }>();

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
    </Page>
  );
}

export default styled(ApplicationDetail)``;
