import * as React from "react";
import styled from "styled-components";
import Alert from "../../components/Alert";
import ErrorList from "../../components/ErrorList";
import NotificationsTable from "../../components/NotificationsTable";
import Page from "../../components/Page";
import { useListProviders } from "../../hooks/notifications";
type Props = {
  className?: string;
};

function Notifications({ className }: Props) {
  const { data, isLoading, error } = useListProviders();
  return (
    <Page
      className={className}
      loading={isLoading}
      path={[{ label: "Notifications" }]}
    >
      {error && (
        <Alert severity="error" title="Request Error" message={error.message} />
      )}
      <ErrorList errors={data?.errors} />
      <NotificationsTable rows={data?.objects} />
    </Page>
  );
}

export default styled(Notifications).attrs({ className: Notifications.name })``;
