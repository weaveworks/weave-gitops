import React from "react";
import Flex from "../../Flex";
import Icon, { IconType } from "../../Icon";
import Text from "../../Text";

interface Props {
  modeName: string;
  showName?: boolean;
}

const capitalizeFirstLetter = (strToCapitalize: string) =>
  strToCapitalize.charAt(0).toUpperCase() + strToCapitalize.slice(1);

const PolicyMode = ({ modeName, showName = false }: Props) => {
  let mode = null;
  switch (modeName.toLocaleLowerCase()) {
    case "audit":
      mode = {
        name: "audit",
        icon: <Icon type={IconType.Policy} color="neutral30" size="large" />,
      };
      break;
    case "admission":
      mode = {
        name: "enforce",
        icon: (
          <Icon type={IconType.VerifiedUser} color="neutral30" size="medium" />
        ),
      };
      break;
    default:
      return <Text>-</Text>;
  }
  return (
    <Flex alignItems="center" start gap="4">
      <span title={!showName ? capitalizeFirstLetter(mode.name) : undefined}>
        {mode.icon}
      </span>
      {showName && (
        <Text capitalize size="medium">
          {mode.name}
        </Text>
      )}
    </Flex>
  );
};

export default PolicyMode;
