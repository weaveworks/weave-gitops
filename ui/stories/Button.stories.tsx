import { Meta } from "@storybook/react";
import React from "react";
import Button from "../components/Button";

export default {
  title: "Button",
  component: Button,
  argTypes: {
    color: {
      description:
        "Color comes from our MuiTheme - see /ui/lib/theme.ts and https://mui.com/customization/theming/",
      type: { name: "string", required: false },
      defaultValue: "primary",
      options: ["primary", "secondary"],
      table: {
        defaultValue: { summary: "primary" },
      },
      control: "radio",
    },
    variant: {
      description: "Pick one of the two MuiButton options",
      type: { name: "string", required: false },
      defaultValue: "outlined",
      options: ["outlined", "contained"],
      table: { defaultValue: { summary: "outlined" } },
      control: "radio",
    },
    loading: {
      description:
        "Changes the Buttons `endIcon` prop to Mui's `<CircularProgress />` and sets `disabled` to `true`",
      type: { name: "boolean", required: false },
      defaultValue: false,
      options: [true, false],
      table: { defaultValue: { summary: "false" } },
      control: "boolean",
    },
    endIcon: {
      description: "`<Icon />` Element to come after `<Button />` content",
      table: {
        type: { summary: "<Icon type={IconType.YOUR_TYPE} size='base' />" },
      },
    },
    onClick: {
      description: "Event Handler",
      action: "click",
    },
    className: {
      description: "CSS MUI Overrides or other styling",
      type: { name: "string", required: false },
    },
  },
} as Meta;

const Template = (args) => <Button {...args}>Weaveworks</Button>;

export const Default = Template.bind({});
// export const AuthButton = Template.bind({});

Default.args = {};
// AuthButton.args = { className: "auth-button", variant: "contained" };
