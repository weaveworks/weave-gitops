import * as React from "react";
import styled from "styled-components";
import KeyValueTable from "../components/KeyValueTable";
import Page from "../components/Page";
import useApplications from "../hooks/applications";
import useNavigation from "../hooks/navigation";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
};

function ApplicationDetail({ className }: Props) {
  const {
    query: { name },
  } = useNavigation<{ name: string }>();
  const { currentApplication: app, getApplication, error } = useApplications();

  React.useEffect(() => {
    getApplication(name);
  }, [name]);

  if (!app) {
    return null;
  }

  return (
    <Page
      error={error}
      loading={false}
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
