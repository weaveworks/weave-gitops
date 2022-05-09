import { DateTime } from "luxon";
import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  time: string;
};

function Timestamp({ className, time }: Props) {
  const t = DateTime.fromISO(time);

  return <span className={className}>{t.toRelative()}</span>;
}

export default styled(Timestamp)``;
