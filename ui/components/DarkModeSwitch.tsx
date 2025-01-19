import { Switch } from "@mui/material";
import * as React from "react";
import styled from "styled-components";
import { AppContext, ThemeTypes } from "../contexts/AppContext";
import images from "../lib/images";
import { svgToB64Image } from "../lib/utils";

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
  .MuiSwitch-switchBase {
    &:hover {
      background-color: ${(props) => props.theme.colors.blueWithOpacity};
    }
  }
  .MuiSwitch-thumb {
    color: #fff;
    background-image: url(${(props) =>
      svgToB64Image(
        props.theme.mode,
        images.lightModeIcon,
        images.darkModeIcon,
      )});
  }
  .MuiSwitch-track {
    background-color: ${(props) => props.theme.colors.primary};
  }
`;
