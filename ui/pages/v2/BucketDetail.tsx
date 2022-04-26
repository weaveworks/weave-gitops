import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";
import BucketDetailComponent from "../../components/BucketDetail";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function BucketDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className} title={name}>
      <BucketDetailComponent
        name={name}
        namespace={namespace}
      />
    </Page>
  );
}

export default styled(BucketDetail).attrs({
  className: BucketDetail.name,
})``;
