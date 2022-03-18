import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";
import SourcesTable from "../../components/SourcesTable";
import { useListSources } from "../../hooks/sources";

type Props = {
  className?: string;
};

function SourcesList({ className }: Props) {
  const { data: sources, error, isLoading, isFetching } = useListSources();
  return (
    <Page
      title="Sources"
      error={error}
      loading={isLoading}
      isFetching={isFetching}
      className={className}
    >
      <SourcesTable sources={sources} />
    </Page>
  );
}

export default styled(SourcesList).attrs({ className: SourcesList.name })`
  h3 {
    visibility: hidden;
  }
`;
