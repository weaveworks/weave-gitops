import * as React from "react";
import styled from "styled-components";
import { removeKind } from "../lib/utils";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { OCIRepository } from "../lib/objects";
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

function OCIRepositoryDetail({
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
      type={FluxObjectKind.KindOCIRepository}
      info={(oci: OCIRepository = new OCIRepository({})) => {
        return [
          ["Type", removeKind(FluxObjectKind.KindOCIRepository)],
          ["URL", <Link href={oci.url}>{oci.url}</Link>],
          [
            "Last Updated",
            oci.lastUpdatedAt ? <Timestamp time={oci.lastUpdatedAt} /> : "-",
          ],
          ["Interval", <Interval interval={oci.interval} />],
          ["Cluster", oci.clusterName],
          ["Namespace", oci.namespace],
          ["Source", <Link href={oci.source}>{oci.source}</Link>],
          ["Revision", oci.revision],
        ];
      }}
    />
  );
}

export default styled(OCIRepositoryDetail).attrs({
  className: OCIRepositoryDetail.name,
})``;
