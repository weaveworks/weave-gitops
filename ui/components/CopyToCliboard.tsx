import * as React from "react";
import { useCallback, useState } from "react";
import Icon, { IconType } from "./Icon";

export default function CopyToClipboard({
  value,
  size,
  showText,
}: {
  value: string;
  className?: string;
  size?: "small" | "base" | "medium" | "large";
  showText?: boolean;
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
  let text = undefined;
  if (showText) {
    text = copied ? "Copied" : "Copy";
  }
  return (
    <div onClick={handleCopy} style={{ cursor: "pointer" }}>
      <Icon
        type={copied ? IconType.CheckMark : IconType.FileCopyIcon}
        size={size}
        text={text}
      />
    </div>
  );
}
