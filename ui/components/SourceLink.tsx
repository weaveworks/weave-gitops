import * as React from "react";
import styled from "styled-components";
import { SourceRef } from "../lib/api/core/types.pb";
import { formatSourceURL } from "../lib/nav";
import Link from "./Link";

type Props = {
  className?: string;
  sourceRef: SourceRef;
};

function SourceLink({ className, sourceRef }: Props) {
  return (
    <Link
      className={className}
      to={formatSourceURL(sourceRef.kind, sourceRef.name)}
    >
      {sourceRef.kind}/{sourceRef.name}
    </Link>
  );
}

export default styled(SourceLink).attrs({ className: SourceLink.name })``;
