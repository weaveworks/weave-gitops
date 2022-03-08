import * as React from "react";
import styled from "styled-components";
import Interval from "../../components/Interval";
import Page, { Content, TitleBar } from "../../components/Page";
import SourceDetail from "../../components/SourceDetail";
import { Bucket, SourceRefSourceKind } from "../../lib/api/core/types.pb";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function BucketDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className}>
      <SourceDetail
        name={name}
        namespace={namespace}
        type={SourceRefSourceKind.Bucket}
        // Guard against an undefined bucket with a default empty object
        info={(b: Bucket = {}) => [
          ["Endpoint", b.endpoint],
          ["Bucket Name", b.name],
          ["Last Updated", ""],
          ["Interval", <Interval interval={b.interval} />],
          ["Cluster", ""],
          ["Namespace", b.namespace],
        ]}
      />
    </Page>
  );
}

export default styled(BucketDetail).attrs({
  className: BucketDetail.name,
})`
  ${TitleBar} {
    margin-bottom: 0;
  }

  ${Content} {
    padding-top: 0;
  }
`;
