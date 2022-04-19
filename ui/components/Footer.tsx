import * as React from "react";
import styled from "styled-components";
import p from "../../package.json";
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

function Footer({ className }: Props) {
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
        {process.env.NODE_ENV !== "test" && (
          <Link
            newTab
            href={`https://github.com/weaveworks/weave-gitops/releases/tag/v${p.version}`}
          >
            v{p.version}
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
