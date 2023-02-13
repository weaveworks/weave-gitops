import * as React from "react";
import { useCallback, useState } from "react";
import styled from "styled-components";
import Icon, { IconType } from "./Icon";
import Text from "./Text";

const PointerIcon = styled(Icon)`
  cursor: pointer;
  svg {
    fill: ${(props) => props.theme.colors.neutral30};
  }
  &.copied > svg {
    fill: ${(props) => props.theme.colors.primary10};
  }
`;

export default function CopyToClipboard({
  value,
  className,
}: {
  value: string;
  className?: string;
}) {
  const [copied, setCopied] = useState(false);
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(value).then(() => {
      setCopied(true);
      setTimeout(() => {
        setCopied(false);
      }, 3000);
    });
  }, [value]);

  return (
    <Text onClick={handleCopy} data-testid="github-code" className={className}>
      <span className="CopyText">{value}</span>
      <PointerIcon
        type={copied ? IconType.CheckMark : IconType.FileCopyIcon}
        className={copied ? "copied" : ""}
        size="medium"
      />
    </Text>
  );
}
