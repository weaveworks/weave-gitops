import * as React from "react";
import styled from "styled-components";
import p from "../../package.json";
import { useVersion } from "../hooks/version";
import { GetVersionResponse } from "../lib/api/core/core.pb";
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
  const versionData = data || ({} as GetVersionResponse);

  const shouldDisplayApiVersion =
    !isLoading &&
    versionData.semver !== p.version &&
    versionData.branch &&
    versionData.commit;

  const versionText = shouldDisplayApiVersion
    ? `${versionData.branch}-${versionData.commit}`
    : `v${p.version}`;
  const versionHref = shouldDisplayApiVersion
    ? `${REPO_URL}/commit/${versionData.commit}`
    : `${REPO_URL}/releases/tag/v${p.version}`;

  const fluxVersionText = versionData.fluxVersion;
  const fluxVersion =
    !isLoading && fluxVersionText !== "" ? (
      <>
        <Text semiBold>Flux:</Text>
        <Spacer padding="xxs" />
        <Text>{fluxVersionText}</Text>
        <Spacer padding="xxs" />
      </>
    ) : null;

  const kubeVersionText = versionData.kubeVersion;
  const kubeVersion =
    !isLoading && kubeVersionText !== "" ? (
      <>
        <Text semiBold>Kubernetes:</Text>
        <Spacer padding="xxs" />
        <Text>{kubeVersionText}</Text>
        <Spacer padding="xxs" />
      </>
    ) : null;

  return (
    <Flex as="footer" wide between className={className} role="footer">
      <LeftFoot>
        <Text>Need help? Contact us at</Text>
        <Spacer padding="xxs" />
        <Link newTab href="mailto:support@weave.works">
          support@weave.works
        </Link>
      </LeftFoot>
      <RightFoot>
        {!isLoading && (
          <>
            <Spacer padding="xxs" />
            {kubeVersion}
            {fluxVersion}
            <Text semiBold noWrap>
              Weave GitOps:
            </Text>
            <Spacer padding="xxs" />
            <Link newTab href={versionHref}>
              <Text semiBold>{versionText}</Text>
            </Link>
          </>
        )}
        <Spacer padding="xxs" />
        <Text>© 2022 Weaveworks</Text>
      </RightFoot>
    </Flex>
  );
}

export default styled(Footer).attrs({ className: Footer.name })`
  color: ${(props) => props.theme.colors.neutral30};
`;
