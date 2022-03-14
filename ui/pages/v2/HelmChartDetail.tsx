import * as React from "react";
import styled from "styled-components";
import Interval from "../../components/Interval";
import Page, { Content, TitleBar } from "../../components/Page";
import SourceDetail from "../../components/SourceDetail";
import Timestamp from "../../components/Timestamp";
import { HelmChart, SourceRefSourceKind } from "../../lib/api/core/types.pb";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function HelmChartDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className}>
      <SourceDetail
        name={name}
        namespace={namespace}
        type={SourceRefSourceKind.HelmChart}
        info={(ch: HelmChart) => [
          ["Chart", ch?.chart],
          ["Ref", ch?.sourceRef?.name],
          ["Last Updated", <Timestamp time={ch?.lastUpdatedAt} />],
          ["Interval", <Interval interval={ch?.interval} />],
          ["Namespace", ch?.namespace],
        ]}
      />
    </Page>
  );
}

export default styled(HelmChartDetail).attrs({
  className: HelmChartDetail.name,
})`
  ${TitleBar} {
    margin-bottom: 0;
  }

  ${Content} {
    padding-top: 0;
  }
`;
