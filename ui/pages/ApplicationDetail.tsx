import * as React from "react";
import styled from "styled-components";
import Icon, { IconType } from "../components/Icon";
import KeyValueTable from "../components/KeyValueTable";
import Page from "../components/Page";
import Timestamp from "../components/Timestamp";
import useNavigation from "../hooks/navigation";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
};

function ApplicationDetail({ className }: Props) {
  const {
    query: { name },
  } = useNavigation<{ name: string }>();

  return (
    <Page
      breadcrumbs={[{ page: PageRoute.Applications }]}
      title={name}
      className={className}
    >
      <KeyValueTable
        columns={4}
        pairs={[
          { key: "name", value: name },
          {
            key: "status",
            value: (
              <Icon
                size="medium"
                color="success"
                type={IconType.CheckMark}
                text="Ready"
              />
            ),
          },
          {
            key: "Last Updated",
            value: <Timestamp time="2006-01-02T15:04:05-0700" />,
          },
        ]}
      />
    </Page>
  );
}

export default styled(ApplicationDetail)``;
