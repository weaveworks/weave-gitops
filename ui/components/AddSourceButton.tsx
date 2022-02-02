import * as React from "react";
import styled from "styled-components";
import { Button } from "..";
import { V2Routes } from "../lib/types";
import { formatURL } from "../lib/utils";
import Link from "./Link";

type Props = {
  className?: string;
  appName?: string;
};

function AddSourceButton({ className, appName }: Props) {
  return (
    <Link className={className} to={formatURL(V2Routes.AddSource, { appName })}>
      <Button>Add Source</Button>
    </Link>
  );
}

export default styled(AddSourceButton).attrs({
  className: AddSourceButton.name,
})``;
