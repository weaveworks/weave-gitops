import { Switch } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { AppContext, ThemeTypes } from "../contexts/AppContext";
import images from "../lib/images";

type Props = {
  className?: string;
  darkModeEnabled?: boolean;
};

function DarkModeSwitch({ className, darkModeEnabled }: Props) {
  const { toggleDarkMode, settings } = React.useContext(AppContext);
  if (!darkModeEnabled) return null;
  return (
    <Switch
      className={className}
      onChange={() => toggleDarkMode()}
      checked={settings.theme === ThemeTypes.Dark}
      color="primary"
    />
  );
}

export default styled(DarkModeSwitch).attrs({
  className: DarkModeSwitch.name,
})`
.MuiSwitch-thumb {
    color: #fff;
    background-image: url(${(props) =>
      props.theme.mode === ThemeTypes.Dark
        ? images.darkModeIcon
        : images.lightModeIcon});
  }
  .MuiSwitch-track {
    background-color: ${(props) => props.theme.colors.primary};
`;
