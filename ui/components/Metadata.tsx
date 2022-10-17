import * as React from "react";
import styled from "styled-components";
import { formatMetadataKey, isHTTP } from "../lib/utils";
import Flex from "./Flex";
import InfoList from "./InfoList";
import Link from "./Link";
import Text from "./Text";

type Props = {
  className?: string;
  metadata?: [string, string][];
  labels?: [string, string][];
};

const Label = styled(Text)`
  border-radius: 10px;
  background-color: ${(props) => props.theme.colors.neutral30};
`;

function Metadata({ metadata, labels, className }: Props) {
  if (!metadata?.length) {
    return <></>;
  }

  const metadataCopy = [];

  for (let i = 0; i < metadata.length; i++) {
    metadataCopy[i] = metadata[i].slice();
  }

  metadataCopy.sort((a, b) => a[0].localeCompare(b[0]));

  metadataCopy.forEach((pair) => {
    pair[0] = formatMetadataKey(pair[0]);

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
      <InfoList items={metadataCopy} />
      <Text size="large" color="neutral30">
        Labels
      </Text>
      <Flex>
        {labels.map((label) => {
          <Label>
            {label[0]}: {label[1]}
          </Label>;
        })}
      </Flex>
    </Flex>
  );
}

export default styled(Metadata).attrs({
  className: Metadata.name,
})`
  margin-top: ${(props) => props.theme.spacing.medium};
`;
