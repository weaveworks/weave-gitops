import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { useListFluxEvents } from "../hooks/events";
import { Event, ObjectRef } from "../lib/api/core/types.pb";
import Alert from "./Alert";
import DataTable from "./DataTable";
import Flex from "./Flex";
import Spacer from "./Spacer";
import Text from "./Text";
import Timestamp from "./Timestamp";

type Props = {
  className?: string;
  namespace?: string;
  involvedObject: ObjectRef;
};

function EventsTable({ className, involvedObject }: Props) {
  const { data, isLoading, error } = useListFluxEvents(involvedObject);

  if (isLoading) {
    return (
      <Flex wide center align>
        <CircularProgress />
      </Flex>
    );
  }

  if (error) {
    return (
      <Spacer padding="small">
        <Alert title="Error" message={error.message} severity="error" />
      </Spacer>
    );
  }

  return (
    <DataTable
      className={className}
      fields={[
        {
          value: (e: Event) => <Text capitalize>{e.reason}</Text>,
          label: "Reason",
        },
        { value: "message", label: "Message", maxWidth: 600 },
        { value: "component", label: "Component" },
        {
          label: "Timestamp",
          value: (e: Event) => <Timestamp time={e.timestamp} />,
        },
      ]}
      rows={data.events}
    />
  );
}

export default styled(EventsTable).attrs({ className: EventsTable.name })`
  td {
    max-width: 1024px;

    &:nth-child(2) {
      white-space: pre-wrap;
      overflow-wrap: break-word;
      word-wrap: break-word;
    }
  }
`;
