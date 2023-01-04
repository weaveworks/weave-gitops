import React from "react";
import Link from "@docusaurus/Link";
import useGlobalData from "@docusaurus/useGlobalData";

const containerStyle = {
  fontSize: 16,
  marginLeft: 4,
  fontVariant: "all-small-caps",
};

export default function TierLabel({ tiers }) {
  return (
    <Link
      title={`This feature is a available on ${tiers}`}
      style={containerStyle}
    >
      {tiers}
    </Link>
  );
}
