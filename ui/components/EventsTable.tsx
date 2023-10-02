import * as React from "react";
import styled from "styled-components";
import { useListEvents } from "../hooks/events";
import { Event, ObjectRef } from "../lib/api/core/types.pb";
import DataTable from "./DataTable";
import Icon, { IconType } from "./Icon";
import RequestStateHandler from "./RequestStateHandler";
import Text from "./Text";
import Timestamp from "./Timestamp";

type Props = {
  className?: string;
  namespace?: string;
  involvedObject: ObjectRef;
};

function EventsTable({ className, involvedObject }: Props) {
  const { data, isLoading, error } = useListEvents(involvedObject);

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <DataTable
        className={className}
        fields={[
          {
            label: "Reason",
            labelRenderer: () => {
              return (
                <h2
                  className="reason"
                  title="This refers to what triggered the event, and can vary by component."
                >
                  Reason
                  <Icon
                    size="base"
                    type={IconType.InfoIcon}
                    color="neutral30"
                  />
                </h2>
              );
            },
            value: (e: Event) => <Text capitalize>{e.reason}</Text>,
            sortValue: (e: Event) => e.reason,
          },
          { label: "Message", value: "message", maxWidth: 600 },
          { label: "From", value: "component" },
          {
            label: "Last Updated",
            value: (e: Event) => <Timestamp time={e.timestamp} />,
            sortValue: (e: Event) => -Date.parse(e.timestamp),
            defaultSort: true,
            secondarySort: true,
          },
        ]}
        rows={data?.events}
      />
    </RequestStateHandler>
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
  .reason {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 16px !important;
  }
`;
