import * as React from "react";
import styled from "styled-components";
import p from "../../package.json";
import { useVersion } from "../hooks/version";
import { GetVersionResponse } from "../lib/api/core/core.pb";
import { getAppVersion } from "../lib/utils";
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

function Footer({ className }: Props) {
  const { data, isLoading } = useVersion();
  const versionData = data || ({} as GetVersionResponse);

  const appVersion = getAppVersion(data, p.version, isLoading, "v");

  const versions: VersionProps[] = !isLoading
    ? [
        {
          productName: "Kubernetes",
          appVersion: {
            versionText: versionData.kubeVersion,
          },
        },
        {
          productName: "Flux",
          appVersion: {
            versionText: versionData.fluxVersion,
          },
        },
        {
          productName: "Weave GitOps",
          appVersion: appVersion,
        },
      ]
    : [];

  return (
    <Flex as="footer" wide between className={className} role="footer">
      <LeftFoot>
        <Text noWrap>Need help? Contact us at</Text>
        <Spacer padding="xxs" />
        <Link newTab href="mailto:sales@weave.works">
          sales@weave.works
        </Link>
      </LeftFoot>
      <RightFoot>
        {!isLoading &&
          versions.map((version, index) => (
            <React.Fragment key={`version ${index}`}>
              <Spacer padding="xxs" />
              <Version {...version} />
            </React.Fragment>
          ))}
        <Spacer padding="xxs" />
        <Text noWrap>© 2022 Weaveworks</Text>
      </RightFoot>
    </Flex>
  );
}

export default styled(Footer).attrs({ className: Footer.name })`
  color: ${(props) => props.theme.colors.neutral30};
  padding-top: ${(props) => props.theme.spacing.small};
`;
