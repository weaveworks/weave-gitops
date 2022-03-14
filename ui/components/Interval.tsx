import * as React from "react";
import styled from "styled-components";
import { Interval as IntervalType } from "../lib/api/core/types.pb";
import { showInterval } from "../lib/time";

type Props = {
  className?: string;
  interval: IntervalType;
};

function Interval({ className, interval }: Props) {
  if (!interval) {
    return null;
  }

  return <span className={className}>{showInterval(interval)}</span>;
}

export default styled(Interval).attrs({ className: Interval.name })``;
