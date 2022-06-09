import * as React from "react";
import styled from "styled-components";
import { isHTTP } from "../lib/utils";
import Flex from "./Flex";
import InfoList from "./InfoList";
import Link from "./Link";
import Text from "./Text";

type Props = {
  className?: string;
  metadata: any;
};

function Metadata({ metadata, className }: Props) {
  if (!metadata?.length) {
    return <></>;
  }

  metadata.forEach((pair) => {
    const data = pair[1];
    if (isHTTP(data))
      pair[1] = (
        <Link newTab href={data}>
          {data}
        </Link>
      );
  });

  return (
    <Flex wide column className={className}>
      <Text size="large" color="neutral30">
        Metadata
      </Text>
      <InfoList items={metadata} />
    </Flex>
  );
}

export default styled(Metadata).attrs({
  className: Metadata.name,
})`
  margin-top: ${(props) => props.theme.spacing.medium};
`;
