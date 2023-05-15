import * as React from "react";
import { useCallback, useState } from "react";
import styled from "styled-components";
import { IconButton } from "./Button";
import Icon, { IconType } from "./Icon";

const CopyButton = styled(IconButton)`
  &.MuiButton-outlinedPrimary {
    margin-left: ${(props) => props.theme.spacing.xxs};
    padding: ${(props) => props.theme.spacing.xxs};
  }
  &.MuiButton-root {
    height: initial;
    width: initial;
    min-width: 0px;
  }
`;

export default function CopyToClipboard({
  value,
  className,
  size,
}: {
  value: string;
  className?: string;
  size?: "small" | "medium" | "large";
}) {
  const [copied, setCopied] = useState(false);
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(value).then(() => {
      setCopied(true);
      setTimeout(() => {
        setCopied(false);
      }, 1500);
    });
  }, [value]);

  return (
    <CopyButton onClick={handleCopy} className={className}>
      <Icon
        type={copied ? IconType.CheckMark : IconType.FileCopyIcon}
        size={size}
      />
    </CopyButton>
  );
}
