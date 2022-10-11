import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { OCIRepository } from "../lib/objects";
import EditButton from "./EditButton";
import { InfoField } from "./InfoList";
import Interval from "./Interval";
import Link from "./Link";
import SourceDetail from "./SourceDetail";
import Timestamp from "./Timestamp";

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
      type={Kind.OCIRepository}
      source={ociRepository}
      customActions={[<EditButton resource={ociRepository} />]}
      info={[
        ["Type", Kind.OCIRepository],
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
