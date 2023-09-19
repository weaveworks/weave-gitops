import React from "react";

import CodeBlock from "@theme/CodeBlock";
import BrowserOnly from "@docusaurus/BrowserOnly";

export default function CurlCodeBlock({ localPath, hostedPath, content }) {
  return (
    <>
      <BrowserOnly>
        {() => (
          <CodeBlock className="language-bash">
            curl -o {localPath} {window.location.protocol}
            //{window.location.host}
            {hostedPath}
          </CodeBlock>
        )}
      </BrowserOnly>

      <CodeBlock title={localPath} className="language-yaml">
        {content}
      </CodeBlock>
    </>
  );
}
