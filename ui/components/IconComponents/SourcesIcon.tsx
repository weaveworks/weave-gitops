import * as React from "react";
import styled from "styled-components";
// eslint-disable-next-line
import { colors } from "../../typedefs/styled";

type Props = {
  className?: string;
  color?: keyof typeof colors;
};

function SourcesIcon({ className }: Props) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      <path
        d="M10.7 14.3L8.4 12L10.7 9.7L10 9L7 12L10 15L10.7 14.3ZM13.3 14.3L15.6 12L13.3 9.7L14 9L17 12L14 15L13.3 14.3V14.3Z"
        fill="#1a1a1a"
      />
      <rect x="4.5" y="4.5" width="15" height="15" rx="7.5" stroke="#1A1A1A" />
    </svg>
  );
}

export default styled(SourcesIcon).attrs({ className: "SourcesIcon" })`
  fill: none !important;
  path {
    fill: ${(props) => props.theme.colors[props.color as any]} !important;
  }
  rect {
    stroke: ${(props) => props.theme.colors[props.color as any]} !important;
    transition: fill 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
  }
`;
