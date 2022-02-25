import * as React from "react";
import styled from "styled-components";
import { Interval as IntervalType } from "../lib/api/core/types.pb";

type Props = {
  className?: string;
  interval: IntervalType;
};

function Interval({ className, interval }: Props) {
  return (
    <span className={className}>
      {interval.hours}h {interval.minutes}m
    </span>
  );
}

export default styled(Interval).attrs({ className: Interval.name })``;
