import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  yaml: string;
}

function YamlView({ yaml, className }: Props) {
  return (
    <pre className={className}>
      <code>
        {yaml}
      </code>
    </pre>
  );
}

export default styled(YamlView).attrs({
  className: YamlView.name,
})`
  font-size: ${props => props.theme.fontSizes.small};
  white-space: break-spaces;
`;
