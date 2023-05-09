import * as React from "react";

function PoliciesIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      strokeLinecap="round"
      strokeMiterlimit={10}
    >
      <g>
        <path
          className="path-fill"
          d="m12,4.09l6,2.58v4.56c-.54,6.18-4.84,8.25-6,8.7-1.16-.45-5.46-2.52-6-8.7v-4.56s6-2.58,6-2.58m0-1.09h0s-7,3.02-7,3.02v5.26c.66,7.85,7,9.71,7,9.71h0s6.34-1.86,7-9.71v-5.26s-7-3.01-7-3.01h0Z"
        />
      </g>
      <g id="explorer">
        <path
          className="path-fill"
          d="m11.98,10.06c.53,0,1.03.21,1.4.58.77.77.77,2.03,0,2.8-.37.37-.87.58-1.4.58s-1.03-.21-1.4-.58c-.77-.77-.77-2.03,0-2.8.37-.37.87-.58,1.4-.58m0-1c-.76,0-1.53.29-2.11.87-1.16,1.16-1.16,3.05,0,4.21.58.58,1.34.87,2.11.87s1.53-.29,2.11-.87c1.16-1.16,1.16-3.05,0-4.21-.58-.58-1.34-.87-2.11-.87h0Z"
        />
        <line
          className="stroke-fill"
          x1="14"
          y1="14.06"
          x2="16.71"
          y2="16.77"
        />
      </g>
    </svg>
  );
}

export default PoliciesIcon;
