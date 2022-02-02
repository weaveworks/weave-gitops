import * as React from "react";
import styled from "styled-components";
import AutomationSelector from "../../components/AutomationSelector";
import Page from "../../components/Page";
import { pageTitleWithAppName } from "../../lib/utils";

type Props = {
  className?: string;
  appName?: string;
};

function AddAutomation({ className, appName }: Props) {
  return (
    <Page
      title={pageTitleWithAppName("Add GitOps Automation", appName)}
      className={className}
    >
      <AutomationSelector appName={appName} />
    </Page>
  );
}

export default styled(AddAutomation).attrs({ className: AddAutomation.name })``;
