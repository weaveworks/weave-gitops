import { Tooltip } from "@material-ui/core";
import { DateTime } from "luxon";
import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  time: string;
  hideSeconds?: boolean;
  tooltip?: boolean;
};

function Timestamp({ className, time, hideSeconds, tooltip }: Props) {
  const dateTime = DateTime.fromISO(time);

  let relativeTime = dateTime.toRelative();
  const fullTime = dateTime.toLocaleString(
    DateTime.DATETIME_SHORT_WITH_SECONDS
  );

  if (hideSeconds && relativeTime.includes("second")) {
    relativeTime = "less than a minute ago";
  }

  if (tooltip)
    return (
      <Tooltip title={fullTime} placement="top">
        <span className={className}>{relativeTime}</span>
      </Tooltip>
    );
  else return <span className={className}>{relativeTime}</span>;
}

export default styled(Timestamp)``;
