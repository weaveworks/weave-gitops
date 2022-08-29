import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";

type Props = {
  className?: string;
};

function Settings({ className }: Props) {
  return <Page className={className}>in progress</Page>;
}

export default styled(Settings).attrs({ className: Settings.name })``;
