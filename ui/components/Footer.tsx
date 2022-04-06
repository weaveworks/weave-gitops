import * as React from "react";
import styled from "styled-components";
import Link from "./Link";

type Props = {
  className?: string;
};

function Footer({ className }: Props) {
  return (
    <footer className={className}>
      Need help? Contact us at{" "}
      <Link newTab href="mailto:support@weave.works">
        support@weave.works
      </Link>
    </footer>
  );
}

export default styled(Footer).attrs({ className: Footer.name })`
  color: ${(props) => props.theme.colors.neutral30};
`;
