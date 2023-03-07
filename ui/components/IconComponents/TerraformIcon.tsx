import * as React from "react";
import styled from "styled-components";
import { colors } from "../../typedefs/styled";

type Props = {
  className?: string;
  color: keyof typeof colors;
};

function TerraformIcon({ className }: Props) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      <g clip-path="url(#clip0_2223_8587)">
        <path
          d="M9.93915 6.83167L14.0997 9.28711V14.1782L9.93915 11.7228V6.83167Z"
          fill="#1A1A1A"
        />
        <path
          d="M14.8385 9.28711L18.9991 6.83167V11.7228L14.8385 14.1782V9.28711Z"
          fill="#1A1A1A"
        />
        <path
          d="M5.00099 4L9.16158 6.45545V11.3465L5.00099 8.89109V4Z"
          fill="#1A1A1A"
        />
        <path
          d="M9.93915 12.6534L14.0997 15.0891V20L9.93915 17.5445V12.6534Z"
          fill="#1A1A1A"
        />
      </g>
      <defs>
        <clipPath id="clip0_2223_8587">
          <rect
            width="14"
            height="16"
            fill="white"
            transform="translate(5 4)"
          />
        </clipPath>
      </defs>
    </svg>
  );
}

export default styled(TerraformIcon).attrs({ className: TerraformIcon.name })`
  path {
    fill: ${(props) => props.theme.colors[props.color as any]} !important;
  }
`;
