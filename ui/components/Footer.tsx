import * as React from "react";
import styled from "styled-components";
import p from "../../package.json";
import Flex from "./Flex";

type Props = {
  className?: string;
  supportComponent?: JSX.Element;
};

function DefaultSupportComponent() {
  return <a href="mailto:sales@weave.works">Commercial Support</a>;
}

function Footer({ className, supportComponent: Support }: Props) {
  if (!Support) {
    Support = <DefaultSupportComponent />;
  }

  return (
    <footer className={className}>
      <Flex wide center>
        <ul>
          <li>© 2021 Weaveworks</li>
          <li>
            <span className="footer__Divider-sc-1vj27u6-5 fNWEGg">•</span>
          </li>
          <li>
            <a href="https://www.weave.works/company/">About</a>
          </li>
          <li>
            <a href={`https://docs.gitops.weave.works/${p.version || ""}`}>
              Docs
            </a>
          </li>
          <li>
            <a href="https://www.weave.works/help/">Help</a>
          </li>
          <li>{Support}</li>
        </ul>
      </Flex>
    </footer>
  );
}

export default styled(Footer).attrs({ className: Footer.name })`
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  font-size: ${(props) => props.theme.fontSizes.small};

  ul {
    list-style: none;
    display: flex;

    li {
      padding: 0 16px;
    }
  }

  a {
    text-decoration: none;
  }
`;
