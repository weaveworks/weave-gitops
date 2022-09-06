import * as React from "react";
import styled from "styled-components";
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
      error={data?.errors || error}
    >
      <NotificationsTable rows={data?.objects} />
    </Page>
  );
}

export default styled(Notifications).attrs({ className: Notifications.name })``;
