import * as React from "react";
import styled from "styled-components";
import Interval from "../../components/Interval";
import Page, { Content, TitleBar } from "../../components/Page";
import SourceDetail from "../../components/SourceDetail";
import {
  HelmRepository,
  SourceRefSourceKind,
} from "../../lib/api/core/types.pb";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function HelmRepositoryDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className}>
      <SourceDetail
        name={name}
        namespace={namespace}
        type={SourceRefSourceKind.HelmRepository}
        // Guard against an undefined bucket with a default empty object
        info={(hr: HelmRepository = {}) => [
          ["URL", hr.url],
          ["Last Updated", ""],
          ["Interval", <Interval interval={hr.interval} />],
          ["Cluster", "Default"],
          ["Namespace", hr.namespace],
        ]}
      />
    </Page>
  );
}

export default styled(HelmRepositoryDetail).attrs({
  className: HelmRepositoryDetail.name,
})`
  ${TitleBar} {
    margin-bottom: 0;
  }

  ${Content} {
    padding-top: 0;
  }
`;
