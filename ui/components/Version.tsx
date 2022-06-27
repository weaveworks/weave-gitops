import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

export type VersionProps = {
  className?: string;
  productName: string;
  versionText?: string;
  versionHref?: string;
};

function Version({
  className,
  productName,
  versionText,
  versionHref,
}: VersionProps) {
  const formattedVersionText = versionText || "-";

  return (
    <Flex className={className} wrap>
      <Text semiBold noWrap>
        {productName}:
      </Text>
      <Spacer padding="xxs" />
      {versionHref && versionText ? (
        <Link newTab href={versionHref}>
          <Text semiBold>{formattedVersionText}</Text>
        </Link>
      ) : (
        <Text>{formattedVersionText}</Text>
      )}
    </Flex>
  );
}

export default styled(Version).attrs({ className: Version.name })``;
