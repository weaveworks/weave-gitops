import * as React from "react";
import styled from "styled-components";
import p from "../../package.json";
import Flex from "./Flex";
import Link from "./Link";

type Props = {
  className?: string;
};

const RightFoot = styled(Flex)`
  padding-right: ${(props) => props.theme.spacing.medium};
  width: 275px;
`;

const LeftFoot = styled(Flex)`
  width: 350px;
`;
function Footer({ className }: Props) {
  return (
    <Flex as="footer" wide between className={className}>
      <LeftFoot>
        Need help? Contact us at
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
        © 2022 Weaveworks
      </RightFoot>
    </Flex>
  );
}

export default styled(Footer).attrs({ className: Footer.name })`
  color: ${(props) => props.theme.colors.neutral30};
`;
