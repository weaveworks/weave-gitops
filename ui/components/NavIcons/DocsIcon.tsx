import * as React from "react";
//<style>.cls-1{fill:#737373;}.cls-2{fill:none;stroke:#737373;stroke-linecap:round;}
function DocsIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
      <path
        className="path-fill"
        d="m19,5v14H5V5h14m0-1H5c-.55,0-1,.45-1,1v14c0,.55.45,1,1,1h14c.55,0,1-.45,1-1V5c0-.55-.45-1-1-1h0Z"
      />
      <line
        className="stroke-fill"
        fill="none"
        x1="8"
        y1="15"
        x2="12"
        y2="15"
      />
      <line
        className="stroke-fill"
        fill="none"
        x1="8"
        y1="12"
        x2="16"
        y2="12"
      />
      <line className="stroke-fill" fill="none" x1="8" y1="9" x2="16" y2="9" />
    </svg>
  );
}

export default DocsIcon;
