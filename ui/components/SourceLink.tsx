import * as React from "react";
import styled from "styled-components";
import { ObjectRef } from "../lib/api/core/types.pb";
import { displayKind } from "../lib/utils";
import { formatSourceURL } from "../lib/nav";
import Link from "./Link";

type Props = {
  className?: string;
  sourceRef?: ObjectRef;
  short?: boolean;
};

function SourceLink({ className, sourceRef, short }: Props) {
  if (!sourceRef) {
    return <div />;
  }
  return (
    <Link
      className={className}
      to={formatSourceURL(sourceRef.kind, sourceRef.name, sourceRef.namespace)}
    >
      {!short && displayKind(sourceRef.kind) + "/"}
      {sourceRef.name}
    </Link>
  );
}

export default styled(SourceLink).attrs({ className: SourceLink.name })``;
