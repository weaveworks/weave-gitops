import React from "react";
import styled from "styled-components";
import { formatURL } from "../lib/nav";
import Link from "./Link";

function extractClusterName(cluster: string, includeNamespace = true): string {
  if (cluster === "management") return "";
  if (includeNamespace) {
    return cluster.split("/")[1];
  }
  return cluster;
}

function ClusterDashboardLink({
  clusterName,
  namespaceIncluded = true,
  clusterDashboardRoute = "/cluster",
}: {
  clusterName: string;
  namespaceIncluded?: boolean;
  clusterDashboardRoute?: string;
}) {
  const clsName = extractClusterName(clusterName || "", namespaceIncluded);
  return (
    <>
      {clsName ? (
        <Link
          to={formatURL(clusterDashboardRoute, {
            clusterName: clsName,
          })}
        >
          {clusterName}
        </Link>
      ) : (
        clusterName
      )}
    </>
  );
}
export default styled(ClusterDashboardLink).attrs({
  className: ClusterDashboardLink.name,
})``;
