import { DateTime } from "luxon";
import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  time: string;
  hideSeconds?: boolean;
};

function Timestamp({ className, time, hideSeconds }: Props) {
  let formattedTime = DateTime.fromISO(time).toRelative();

  if (hideSeconds && formattedTime.includes("second")) {
    formattedTime = "less than a minute ago";
  }

  return <span className={className}>{formattedTime}</span>;
}

export default styled(Timestamp)``;
