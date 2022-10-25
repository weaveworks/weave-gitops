import * as React from "react";
import styled from "styled-components";
import { formatMetadataKey, isHTTP } from "../lib/utils";
import Flex from "./Flex";
import InfoList from "./InfoList";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

type Props = {
  className?: string;
  metadata?: [string, string][];
  labels?: [string, string][];
};

const Label = styled(Text)`
  padding: ${(props) => props.theme.spacing.xs}
    ${(props) => props.theme.spacing.small};
  margin-right: ${(props) => props.theme.spacing.xxs};
  border-radius: 15px;
  white-space: nowrap;
  background-color: ${(props) => props.theme.colors.neutral20};
`;

const LabelFlex = styled(Flex)`
  padding: ${(props) => props.theme.spacing.xxs} 0;
  overflow-x: scroll;
`;

function Metadata({ metadata = [], labels = [], className }: Props) {
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
      {metadataCopy.length > 0 && (
        <>
          <Text size="large" color="neutral30">
            Metadata
          </Text>
          <InfoList items={metadataCopy} />
          <Spacer padding="small" />
        </>
      )}
      {labels.length > 0 && (
        <>
          <Text size="large" color="neutral30">
            Labels
          </Text>
          <LabelFlex wide start>
            {labels.map((label, index) => {
              return (
                <Label key={index}>
                  {label[0]}: {label[1]}
                </Label>
              );
            })}
          </LabelFlex>
          <Spacer padding="small" />
        </>
      )}
    </Flex>
  );
}

export default styled(Metadata).attrs({
  className: Metadata.name,
})`
  margin-top: ${(props) => props.theme.spacing.medium};
`;
