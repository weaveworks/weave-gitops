import React from "react";
import Flex from "../../Flex";
import Icon, { IconType } from "../../Icon";
import Text from "../../Text";

const Severity = ({ severity }: { severity: string }) => {
  let icon = null;
  switch (severity.toLocaleLowerCase()) {
    case "low":
      icon = (
        <Icon type={IconType.CallReceived} color="primary20" size="base" />
      );
      break;
    case "medium":
      icon = <Icon type={IconType.Remove} color="feedbackDark" size="base" />;
      break;
    case "high":
      icon = <Icon type={IconType.CallMade} color="alertDark" size="base" />;
      break;
  }
  return (
    <Flex alignItems="center" gap="4" data-testid={severity}>
      {icon}
      <Text capitalize size="medium">
        {severity}
      </Text>
    </Flex>
  );
};

export default Severity;
