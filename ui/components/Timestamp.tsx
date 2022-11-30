import { DateTime } from "luxon";
import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  time: string;
  duration?: boolean;
};

function Timestamp({ className, time, duration }: Props) {
  const t = DateTime.fromISO(time);
  let formatted;
  if (duration) {
    formatted = t.diffNow();
    formatted = formatted.toHuman();
  } else formatted = t.toRelative();
  return <span className={className}>{formatted}</span>;
}

export default styled(Timestamp)``;
