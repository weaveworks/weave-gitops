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
  artifactMetadata?: [string, string][];
  labels?: [string, string][];
};

const Label = styled(Text)`
  padding: ${(props) => props.theme.spacing.xs}
    ${(props) => props.theme.spacing.small};
  border-radius: 15px;
  white-space: nowrap;
  background-color: ${(props) => props.theme.colors.neutralGray};
`;

const makeMetadata = (metadata: [string, string][]): [string, any][] => {
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
  return metadataCopy;
};

function Metadata({
  metadata = [],
  labels = [],
  artifactMetadata = [],
  className,
}: Props) {
  const metadataCopy = makeMetadata(metadata);
  const artifactMetadataCopy = makeMetadata(artifactMetadata);

  return (
    <Flex wide column className={className} gap="12">
      {metadataCopy.length > 0 && (
        <Flex column gap="8">
          <Text size="large" color="neutral30">
            Metadata
          </Text>
          <InfoList items={metadataCopy} />
        </Flex>
      )}
      {artifactMetadataCopy.length > 0 && (
        <Flex column gap="8">
          <Text size="large" color="neutral30">
            Artifact Metadata
          </Text>
          <InfoList items={artifactMetadataCopy} />
        </Flex>
      )}
      {labels.length > 0 && (
        <Flex column gap="8">
          <Text size="large" color="neutral30">
            Labels
          </Text>
          <Flex wide start wrap gap="4">
            {labels.map((label, index) => {
              return (
                <Label key={index}>
                  {label[0]}: {label[1]}
                </Label>
              );
            })}
          </Flex>
        </Flex>
      )}
    </Flex>
  );
}

export default styled(Metadata).attrs({
  className: Metadata.name,
})`
  margin: ${(props) => props.theme.spacing.small} 0;
`;
