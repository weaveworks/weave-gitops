import * as React from "react";
import styled from "styled-components";
import p from "../../package.json";
import { useVersion } from "../hooks/version";
import { GetVersionResponse } from "../lib/api/core/core.pb";
import Flex from "./Flex";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";
import Version, { VersionProps } from "./Version";

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
  const versionData = data || ({} as GetVersionResponse);

  const shouldDisplayApiVersion =
    !isLoading &&
    versionData.semver !== p.version &&
    versionData.branch &&
    versionData.commit;

  const wegoVersionText = shouldDisplayApiVersion
    ? `${versionData.branch}-${versionData.commit}`
    : `v${p.version}`;
  const wegoVersionHref = shouldDisplayApiVersion
    ? `${REPO_URL}/commit/${versionData.commit}`
    : `${REPO_URL}/releases/tag/v${p.version}`;

  const versions: VersionProps[] = !isLoading
    ? [
        {
          productName: "Kubernetes",
          versionText: versionData.kubeVersion,
        },
        {
          productName: "Flux",
          versionText: versionData.fluxVersion,
        },
        {
          productName: "Weave GitOps",
          versionText: wegoVersionText,
          versionHref: wegoVersionHref,
        },
      ]
    : [];

  return (
    <Flex as="footer" wide between className={className} role="footer">
      <LeftFoot>
        <Text noWrap>Need help? Contact us at</Text>
        <Spacer padding="xxs" />
        <Link newTab href="mailto:support@weave.works">
          support@weave.works
        </Link>
      </LeftFoot>
      <RightFoot>
        {!isLoading &&
          versions.map((version, index) => (
            <>
              <Spacer padding="xxs" />
              <Version key={`key ${index}`} {...version} />
            </>
          ))}
        <Spacer padding="xxs" />
        <Text noWrap>© 2022 Weaveworks</Text>
      </RightFoot>
    </Flex>
  );
}

export default styled(Footer).attrs({ className: Footer.name })`
  color: ${(props) => props.theme.colors.neutral30};
`;
