import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

export const repoUrl = "https://github.com/weaveworks/weave-gitops";

export const noVersion = "no version";

export type AppVersion = {
  versionText?: string;
  versionHref?: string;
};

export type VersionProps = {
  className?: string;
  productName: string;
  appVersion?: AppVersion;
};

function Version({ className, productName, appVersion }: VersionProps) {
  const formattedVersionText = appVersion?.versionText || "-";

  return (
    <Flex className={className}>
      <Text semiBold noWrap>
        {productName}:
      </Text>
      <Spacer padding="xxs" />
      {appVersion?.versionHref && appVersion?.versionText ? (
        <Link newTab href={appVersion?.versionHref}>
          <Text semiBold>{formattedVersionText}</Text>
        </Link>
      ) : (
        <Text>{formattedVersionText}</Text>
      )}
    </Flex>
  );
}

export default styled(Version).attrs({ className: Version.name })``;
