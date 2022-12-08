import React from "react";
import styled from "styled-components";
import { useLinkResolver } from "../contexts/LinkResolverContext";
import Link from "./Link";
import Text from "./Text";

function ClusterDashboardLink({ clusterName }: { clusterName: string }) {
  const resolver = useLinkResolver();
  const resolved = resolver && resolver("ClusterDashboard", { clusterName });

  if (resolved && clusterName != "management") {
    return <Link to={resolved}>{clusterName}</Link>;
  }

  return <Text>{clusterName}</Text>;
}
export default styled(ClusterDashboardLink).attrs({
  className: ClusterDashboardLink.name,
})``;
