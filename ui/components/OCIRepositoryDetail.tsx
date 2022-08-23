import * as React from "react";
import styled from "styled-components";
import { removeKind } from "../lib/utils";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { OCIRepository } from "../lib/objects";
import { useFeatureFlags } from "../hooks/featureflags";
import Interval from "./Interval";
import Link from "./Link";
import SourceDetail from "./SourceDetail";
import Timestamp from "./Timestamp";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  ociRepository: OCIRepository;
};

function OCIRepositoryDetail({ className, ociRepository }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  const tenancyInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && ociRepository.tenant
      ? [["Tenant", ociRepository.tenant]]
      : [];
  const clusterInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [["Cluster", ociRepository.clusterName]]
      : [];

  return (
    <SourceDetail
      className={className}
      type={FluxObjectKind.KindOCIRepository}
      source={ociRepository}
      info={[
        ["Type", removeKind(FluxObjectKind.KindOCIRepository)],
        ["URL", <Link href={ociRepository.url}>{ociRepository.url}</Link>],
        [
          "Last Updated",
          ociRepository.lastUpdatedAt ? (
            <Timestamp time={ociRepository.lastUpdatedAt} />
          ) : (
            "-"
          ),
        ],
        ["Interval", <Interval interval={ociRepository.interval} />],
        ...clusterInfo,
        ["Namespace", ociRepository.namespace],
        [
          "Source",
          <Link href={ociRepository.source}>{ociRepository.source}</Link>,
        ],
        ["Revision", ociRepository.revision],
        ...tenancyInfo,
      ]}
    />
  );
}

export default styled(OCIRepositoryDetail).attrs({
  className: OCIRepositoryDetail.name,
})``;
