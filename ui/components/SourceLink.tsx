import * as React from "react";
import styled from "styled-components";
import { FluxObjectRef } from "../lib/api/core/types.pb";
import { removeKind } from "../lib/utils";
import { formatSourceURL } from "../lib/nav";
import Link from "./Link";

type Props = {
  className?: string;
  sourceRef?: FluxObjectRef;
  clusterName: string;
  short?: boolean;
};

function SourceLink({ className, sourceRef, clusterName, short }: Props) {
  if (!sourceRef) {
    return <div />;
  }
  return (
    <Link
      className={className}
      to={formatSourceURL(
        sourceRef.kind,
        sourceRef.name,
        sourceRef.namespace,
        clusterName
      )}
    >
      {!short && removeKind(sourceRef.kind) + "/"}
      {sourceRef.name}
    </Link>
  );
}

export default styled(SourceLink).attrs({ className: SourceLink.name })``;
