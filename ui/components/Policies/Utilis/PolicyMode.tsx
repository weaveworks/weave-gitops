import React from "react";
import Icon, { IconType } from "../../Icon";
import { ModeWrapper } from "./PolicyUtilis";

interface IModeProps {
  modeName: string;
  showName?: boolean;
}
const capitalizeFirstLetter = (strToCapitalize: string) =>
  strToCapitalize.charAt(0).toUpperCase() + strToCapitalize.slice(1);

function PolicyMode({ modeName, showName = false }: IModeProps) {
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
}

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
