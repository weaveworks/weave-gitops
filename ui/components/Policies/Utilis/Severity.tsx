import React from "react";
import Flex from "../../Flex";
import Icon, { IconType } from "../../Icon";
import Text from "../../Text";

function Severity({ severity }: { severity: string }) {
  return (
    <Flex alignItems="center" gap="4" data-testid={severity}>
      {(() => {
        switch (severity.toLocaleLowerCase()) {
          case "low":
            return (
              <Icon
                type={IconType.CallReceived}
                color="primary20"
                size="base"
              />
            );
          case "medium":
            return (
              <Icon type={IconType.Remove} color="feedbackDark" size="base" />
            );
          case "high":
            return (
              <Icon type={IconType.CallMade} color="alertDark" size="base" />
            );
        }
      })()}
      <Text capitalize size="medium">
        {severity}
      </Text>
    </Flex>
  );
}

export default Severity;