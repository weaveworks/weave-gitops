import * as React from "react";

function NotificationsIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
      <g id="notifications">
        <polyline className="path-fill" points="4 19.98 4 16.98 7 16.98" />
        <rect className="stroke-fill" x="4" y="16" width="2.88" height=".99" />
      </g>
      <g id="gitops_run">
        <path
          className="path-fill"
          d="m19,5v11H5V5h14m0-1H5c-.55,0-1,.45-1,1v11c0,.55.45,1,1,1h14c.55,0,1-.45,1-1V5c0-.55-.45-1-1-1h0Z"
        />
      </g>
      <g id="policy_config">
        <line className="stroke-fill" x1="12" y1="7.49" x2="12" y2="11.49" />
        <rect className="path-fill" x="11.5" y="13.01" width="1" height="1" />
      </g>
    </svg>
  );
}

export default NotificationsIcon;
