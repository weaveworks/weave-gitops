import * as React from "react";
import styled from "styled-components";
import { removeKind } from "../lib/utils";
import { FluxObjectKind, HelmRepository } from "../lib/api/core/types.pb";
import Interval from "./Interval";
import Link from "./Link";
import SourceDetail from "./SourceDetail";
import Timestamp from "./Timestamp";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function HelmRepositoryDetail({
  name,
  namespace,
  className,
  clusterName,
}: Props) {
  return (
    <SourceDetail
      className={className}
      name={name}
      namespace={namespace}
      clusterName={clusterName}
      type={FluxObjectKind.KindHelmRepository}
      // Guard against an undefined repo with a default empty object
      info={(hr: HelmRepository = {}) => [
        ["Type", removeKind(FluxObjectKind.KindHelmRepository)],
        ["Repository Type", hr.repositoryType.toLowerCase()],
        ["URL", <Link>hr.url</Link>],
        [
          "Last Updated",
          hr.lastUpdatedAt ? <Timestamp time={hr.lastUpdatedAt} /> : "-",
        ],
        ["Interval", <Interval interval={hr.interval} />],
        ["Cluster", hr.clusterName],
        ["Namespace", hr.namespace],
      ]}
    />
  );
}

export default styled(HelmRepositoryDetail).attrs({
  className: HelmRepositoryDetail.name,
})``;
