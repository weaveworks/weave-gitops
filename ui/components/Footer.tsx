import * as React from "react";
import styled from "styled-components";
import p from "../../package.json";
import Flex from "./Flex";

type Props = {
  className?: string;
  extraListItems?: JSX.Element;
};

function DefaultExtraLinks() {
  return (
    <>
      <li>
        <a target="_blank" href="mailto:sales@weave.works">
          Commercial Support
        </a>
      </li>
      <li>
        <a
          style={{ justifySelf: "flex-end" }}
          target="_blank"
          href={`https://github.com/weaveworks/weave-gitops/releases/tag/v${p.version}`}
        >
          v{p.version}
        </a>
      </li>
    </>
  );
}

function Footer({ className, extraListItems }: Props) {
  if (!extraListItems) {
    extraListItems = <DefaultExtraLinks />;
  }

  return (
    <footer className={className}>
      <Flex align wide center>
        <ul>
          <li>© 2021 Weaveworks</li>
          <li>
            <span>•</span>
          </li>
          <li>
            <a
              target="_blank"
              href="https://www.weave.works/product/gitops-core/"
            >
              About
            </a>
          </li>
          <li>
            <a
              target="_blank"
              href={`https://docs.gitops.weave.works/${p.version || ""}`}
            >
              Docs
            </a>
          </li>
          <li>
            <a target="_blank" href="https://www.weave.works/help/">
              Help
            </a>
          </li>
          {extraListItems}
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
