import { Meta, Story } from "@storybook/react";
import React from "react";
import Button from "../components/Button";

export default {
  title: "Button",
  component: Button,
  parameters: {
    docs: {
      description: {
        component:
          "MUI Material Button - see https://mui.com/components/buttons/",
      },
    },
  },
  argTypes: {
    color: {
      description:
        "The three options come from our `MuiTheme` - see /ui/lib/theme.ts and https://mui.com/customization/theming/",
      type: { summary: "string", required: false },
      defaultValue: "primary",
      options: ["primary", "secondary", "inherit"],
      table: {
        defaultValue: { summary: "primary" },
      },
      control: "radio",
    },
    variant: {
      description: "Pick one of the two MuiButton options",
      type: { summary: "string", required: false },
      defaultValue: "outlined",
      options: ["outlined", "contained"],
      table: { defaultValue: { summary: "outlined" } },
      control: "radio",
    },
    loading: {
      description:
        "Changes the Buttons `endIcon` prop to Mui's `<CircularProgress />` and sets `disabled` to `true`",
      type: { summary: "boolean", required: false },
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
      type: { summary: "function(e) {}", name: "function" },
      action: "click",
    },
    className: {
      description: "CSS MUI Overrides or other styling",
      type: { summary: "string", name: "string", required: false },
    },
  },
} as Meta;

const Template: Story = (args) => <Button {...args}>Weaveworks</Button>;

export const Default = Template.bind({});
// export const AuthButton = Template.bind({});
export const ModalButton = Template.bind({});
ModalButton.args = {
  color: "inherit",
  className: "borderless",
};

Default.args = {};
// AuthButton.args = { className: "auth-button", variant: "contained" };
