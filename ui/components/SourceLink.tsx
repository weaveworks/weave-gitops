import * as React from "react";
import styled from "styled-components";
import { SourceRef } from "../lib/api/core/types.pb";
import { formatSourceURL } from "../lib/nav";
import Link from "./Link";

type Props = {
  className?: string;
  sourceRef?: SourceRef;
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
      {!short && sourceRef.kind}/{sourceRef.name}
    </Link>
  );
}

export default styled(SourceLink).attrs({ className: SourceLink.name })``;
