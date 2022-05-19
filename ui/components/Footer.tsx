import * as React from "react";
import styled from "styled-components";
import p from "../../package.json";
import { useVersion, Version } from "../hooks/version";
import Flex from "./Flex";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

type Props = {
  className?: string;
};

const RightFoot = styled(Flex)`
  padding-right: ${(props) => props.theme.spacing.medium};
`;

const LeftFoot = styled(Flex)``;

const REPO_URL = "https://github.com/weaveworks/weave-gitops";

function Footer({ className }: Props) {
  const { data, isLoading } = useVersion();
  const versionData = data?.version || ({} as Version);
  const apiVersion: Version = {
    version: versionData.version,
    gitCommit: versionData["git-commit"],
    branch: versionData.branch,
    buildTime: versionData["buildtime"],
  };

  const shouldDisplayApiVersion =
    !isLoading &&
    apiVersion.version !== p.version &&
    apiVersion.branch &&
    apiVersion.gitCommit;

  const versionText = shouldDisplayApiVersion
    ? `${apiVersion.branch}-${apiVersion.gitCommit}`
    : `v${p.version}`;
  const versionHref = shouldDisplayApiVersion
    ? `${REPO_URL}/commit/${apiVersion.gitCommit}`
    : `${REPO_URL}/releases/tag/v${p.version}`;

  return (
    <Flex as="footer" wide between className={className}>
      <LeftFoot>
        <Text color="neutral30">Need help? Contact us at</Text>
        <Spacer padding="xxs" />
        <Link newTab href="mailto:support@weave.works">
          support@weave.works
        </Link>
      </LeftFoot>
      <RightFoot>
        {process.env.NODE_ENV !== "test" && !isLoading && (
          <Link newTab href={versionHref}>
            {versionText}
          </Link>
        )}
        <Spacer padding="xxs" />
        <Text color="neutral30">© 2022 Weaveworks</Text>
      </RightFoot>
    </Flex>
  );
}

export default styled(Footer).attrs({ className: Footer.name })`
  color: ${(props) => props.theme.colors.neutral30};
`;
