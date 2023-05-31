import React from "react";
import styled from "styled-components";
import Flex from "../../Flex";
import Icon, { IconType } from "../../Icon";

const ModeWrapper = styled(Flex)`
  align-items: center;
  justify-content: flex-start;
  display: inline-flex;
  margin-right: ${(props) => props.theme.spacing.xs};
  svg {
    color: ${(props) => props.theme.colors.neutral30};
    font-size: ${(props) => props.theme.fontSizes.large};
    margin-right: ${(props) => props.theme.spacing.xxs};
  }
  span {
    text-transform: capitalize;
  }
`;
interface Props {
  modeName: string;
  showName?: boolean;
}

const capitalizeFirstLetter = (strToCapitalize: string) =>
  strToCapitalize.charAt(0).toUpperCase() + strToCapitalize.slice(1);

const PolicyMode = ({ modeName, showName = false }: Props) => {
  switch (modeName.toLocaleLowerCase()) {
    case "audit":
      return ModeTooltip(
        "audit",
        showName,
        <Icon size="medium" type={IconType.Policy} color="neutral30" />
      );
    case "admission":
      return ModeTooltip(
        "enforce",
        showName,
        <Icon size="medium" type={IconType.VerifiedUser} color="neutral30" />
      );
    default:
      return (
        <ModeWrapper>
          <span>-</span>
        </ModeWrapper>
      );
  }
};

const ModeTooltip = (mode: string, showName: boolean, icon: any) => {
  return (
    <>
      {!showName ? (
        <ModeWrapper>
          <span title={capitalizeFirstLetter(mode)}>{icon}</span>
        </ModeWrapper>
      ) : (
        <ModeWrapper>
          {icon}
          <span>{mode}</span>
        </ModeWrapper>
      )}
    </>
  );
};

export default PolicyMode;
