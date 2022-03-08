import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { useListFluxEvents } from "../hooks/events";
import { ObjectReference } from "../lib/api/core/types.pb";
import { WeGONamespace } from "../lib/types";
import Alert from "./Alert";
import DataTable from "./DataTable";
import Flex from "./Flex";
import Spacer from "./Spacer";

type Props = {
  className?: string;
  namespace?: string;
  involvedObject: ObjectReference;
};

function EventsTable({
  className,
  namespace = WeGONamespace,
  involvedObject,
}: Props) {
  const { data, isLoading, error } = useListFluxEvents(
    namespace,
    involvedObject
  );

  if (isLoading) {
    return (
      <Flex wide align>
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
        { value: "message", label: "Message" },
        { value: "source", label: "Source" },
        { value: "reason", label: "Reason" },
      ]}
      rows={data.events}
    />
  );
}

export default styled(EventsTable).attrs({ className: EventsTable.name })``;
