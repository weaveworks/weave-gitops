import * as React from "react";
import styled from "styled-components";
import Text from "./Text";
import Flex from "./Flex";
import InfoList from "./InfoList";

type Props = {
  className?: string;
  metadata: any;
};

function Metadata({ metadata, className }: Props) {
  if (!metadata?.length) {
    return <></>;
  }

  return (
    <Flex wide column className={className}>
      <Text size="large">Metadata</Text>
      <InfoList items={metadata} />
    </Flex>
  );
}

export default styled(Metadata).attrs({
  className: Metadata.name,
})`
  margin-top: ${(props) => props.theme.spacing.medium};
`;
